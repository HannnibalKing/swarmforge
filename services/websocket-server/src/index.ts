import { WebSocketServer, WebSocket } from 'ws';
import { createClient } from 'redis';
import http from 'http';

const PORT = process.env.WS_PORT || 3002;
const REDIS_URL = process.env.REDIS_URL || 'redis://localhost:6379';

interface Client {
  ws: WebSocket;
  userId: string;
  subscriptions: Set<string>;
}

const clients = new Map<string, Client>();

// Redis pub/sub for distributed real-time updates
const redisSubscriber = createClient({ url: REDIS_URL });
const redisPublisher = createClient({ url: REDIS_URL });

async function setupRedis() {
  await redisSubscriber.connect();
  await redisPublisher.connect();
  
  // Subscribe to all job and printer updates
  await redisSubscriber.subscribe('job:*', (message: string, channel: string) => {
    broadcastToSubscribers(channel, message);
  });
  
  await redisSubscriber.subscribe('printer:*', (message: string, channel: string) => {
    broadcastToSubscribers(channel, message);
  });
  
  console.log('Redis pub/sub connected');
}

function broadcastToSubscribers(channel: string, message: string) {
  const data = JSON.parse(message);
  
  for (const [clientId, client] of clients) {
    if (client.subscriptions.has(channel) || client.subscriptions.has('*')) {
      if (client.ws.readyState === WebSocket.OPEN) {
        client.ws.send(JSON.stringify({
          type: 'update',
          channel,
          data,
        }));
      }
    }
  }
}

// Create HTTP server and WebSocket server
const server = http.createServer();
const wss = new WebSocketServer({ server });

wss.on('connection', (ws: WebSocket, req: http.IncomingMessage) => {
  const clientId = generateClientId();
  
  console.log(`Client connected: ${clientId}`);
  
  const client: Client = {
    ws,
    userId: '', // Set after authentication
    subscriptions: new Set(),
  };
  
  clients.set(clientId, client);
  
  // Send welcome message
  ws.send(JSON.stringify({
    type: 'connected',
    clientId,
    message: 'Connected to SwarmForge WebSocket',
  }));
  
  ws.on('message', async (message: string) => {
    try {
      const data = JSON.parse(message.toString());
      
      switch (data.type) {
        case 'auth':
          // Validate JWT token
          client.userId = data.userId; // After validation
          ws.send(JSON.stringify({ type: 'auth_success', userId: client.userId }));
          break;
          
        case 'subscribe':
          // Subscribe to specific channels
          // e.g., 'job:123', 'printer:456', 'user:789'
          const channels = Array.isArray(data.channels) ? data.channels : [data.channels];
          channels.forEach((channel: string) => client.subscriptions.add(channel));
          ws.send(JSON.stringify({ 
            type: 'subscribed', 
            channels,
          }));
          break;
          
        case 'unsubscribe':
          const unsubChannels = Array.isArray(data.channels) ? data.channels : [data.channels];
          unsubChannels.forEach((channel: string) => client.subscriptions.delete(channel));
          ws.send(JSON.stringify({ 
            type: 'unsubscribed', 
            channels: unsubChannels,
          }));
          break;
          
        case 'ping':
          ws.send(JSON.stringify({ type: 'pong' }));
          break;
          
        default:
          console.log('Unknown message type:', data.type);
      }
    } catch (error) {
      console.error('Message handling error:', error);
      ws.send(JSON.stringify({ type: 'error', message: 'Invalid message format' }));
    }
  });
  
  ws.on('close', () => {
    console.log(`Client disconnected: ${clientId}`);
    clients.delete(clientId);
  });
  
  ws.on('error', (error) => {
    console.error(`WebSocket error for client ${clientId}:`, error);
  });
});

function generateClientId(): string {
  return `client_${Date.now()}_${Math.random().toString(36).substring(7)}`;
}

// Publish update (would be called from other services via Redis)
export async function publishUpdate(channel: string, data: any) {
  await redisPublisher.publish(channel, JSON.stringify(data));
}

// Startup
async function start() {
  await setupRedis();
  
  server.listen(PORT, () => {
    console.log(`WebSocket server running on port ${PORT}`);
  });
}

start().catch(console.error);
