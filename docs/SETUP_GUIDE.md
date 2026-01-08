# SwarmForge Setup & Deployment Guide

## Quick Start (Development)

### Prerequisites
- Docker Desktop installed
- Git
- 8GB+ RAM available
- 20GB+ disk space

### Installation

1. **Clone repository**
```bash
git clone <repository-url>
cd Proj-Swarmforge
```

2. **Environment setup**
```bash
cp .env.example .env
# Edit .env with your settings (optional for local dev)
```

3. **Start all services**
```bash
docker-compose up -d
```

4. **Initialize database**
```bash
# Run migrations
docker-compose exec postgres psql -U swarmforge -d swarmforge -f /docker-entrypoint-initdb.d/01_schema.sql
```

5. **Setup MinIO buckets**
```bash
make setup-buckets
```

6. **Access services**
- Frontend: http://localhost:3000
- API Gateway: http://localhost:8080
- MinIO Console: http://localhost:9001 (minioadmin/minioadmin)
- RabbitMQ Management: http://localhost:15672 (swarmforge/dev_password)

## Development Workflow

### Run individual services

**Frontend development:**
```bash
cd frontend
npm install
npm run dev
```

**Go services with hot reload:**
```bash
cd services/api-gateway
go install github.com/cosmtrek/air@latest
air
```

**QA Service development:**
```bash
cd services/qa-service
npm install
npm run dev
```

### Database migrations
```bash
# Create new migration
psql -U swarmforge -d swarmforge -h localhost

# Apply migrations
make migrate
```

### Testing

**Run all tests:**
```bash
make test
```

**Individual service tests:**
```bash
cd services/api-gateway
go test ./...

cd services/qa-service
npm test
```

## Production Deployment

### Option 1: Docker Swarm

1. **Initialize swarm**
```bash
docker swarm init
```

2. **Deploy stack**
```bash
docker stack deploy -c docker-compose.prod.yml swarmforge
```

3. **Scale services**
```bash
docker service scale swarmforge_api-gateway=3
docker service scale swarmforge_job-orchestrator=2
```

### Option 2: Kubernetes

1. **Create namespace**
```bash
kubectl create namespace swarmforge
```

2. **Create secrets**
```bash
kubectl create secret generic swarmforge-secrets \
  --from-literal=db-password=<secure-password> \
  --from-literal=jwt-secret=<secure-secret> \
  -n swarmforge
```

3. **Deploy services**
```bash
kubectl apply -f k8s/
```

4. **Check status**
```bash
kubectl get pods -n swarmforge
kubectl get services -n swarmforge
```

### Environment Variables (Production)

```bash
# Database
DATABASE_URL=postgres://user:password@db-host:5432/swarmforge?sslmode=require
DB_POOL_SIZE=20

# Redis
REDIS_URL=redis-host:6379
REDIS_PASSWORD=<secure-password>

# JWT
JWT_SECRET=<strong-random-secret-256-bit>
JWT_EXPIRY=24h

# S3/MinIO
S3_ENDPOINT=https://s3.your-domain.com
S3_ACCESS_KEY=<access-key>
S3_SECRET_KEY=<secret-key>
S3_USE_SSL=true

# External Services
AI_MODEL_ENDPOINT=https://ai.your-domain.com
PAYMENT_GATEWAY_URL=https://payments.your-domain.com

# Monitoring
SENTRY_DSN=<sentry-dsn>
PROMETHEUS_ENDPOINT=http://prometheus:9090
```

## Printer Onboarding Guide

### For Bambu Lab Printers

1. **Register printer**
   - Login to SwarmForge
   - Navigate to "My Printers" → "Add Printer"
   - Select manufacturer: "Bambu Lab"
   - Select model: X1C, P1S, P1P, or A1
   - Fill in details:
     - Printer nickname
     - Location (city, country)
     - Supported materials
     - Upload photo of printer

2. **Run certification tests**
   
   **For all tiers:**
   - Download calibration cube STL
   - Print with these settings:
     - Material: PLA
     - Layer height: 0.2mm
     - Infill: 20%
     - Speed: Standard
   - Let cool 30+ minutes
   - Measure with calipers (X, Y, Z)
   - Upload 4 photos (front, back, top, detail)
   - Submit measurements

   **For Gold/Platinum:**
   - Complete additional tests:
     - Dimensional accuracy test
     - Bridging test (Platinum only)
     - Overhang test (Platinum only)

3. **Certification review**
   - QA team reviews within 24-48 hours
   - Tier assigned based on:
     - Model default tier
     - Calibration test scores
     - Dimensional accuracy
   - Receive notification when approved

4. **Start accepting jobs**
   - Set availability status
   - Configure pricing preferences
   - Enable notifications
   - Accept job assignments

### Expected Certification Results

**Bambu X1 Carbon:**
- Default tier: Platinum
- Expected tolerance: ±0.03-0.05mm
- Certification score: 95-100
- All tests required

**Bambu P1S:**
- Default tier: Gold
- Expected tolerance: ±0.05-0.08mm
- Certification score: 90-95
- Calibration + Dimensional tests

**Bambu P1P:**
- Default tier: Gold
- Expected tolerance: ±0.05-0.08mm
- Certification score: 88-93
- Calibration + Dimensional tests

**Bambu A1:**
- Default tier: Silver
- Expected tolerance: ±0.08-0.10mm
- Certification score: 85-90
- Calibration test only

### For Prusa Printers

**Prusa XL:**
- Default tier: Platinum
- Expected tolerance: ±0.03-0.05mm
- Certification score: 95-100
- All tests required

**Prusa MK4:**
- Default tier: Gold
- Expected tolerance: ±0.05-0.07mm
- Certification score: 92-96
- Calibration + Dimensional + Bridging

**Prusa MK3S+:**
- Default tier: Gold
- Expected tolerance: ±0.05-0.08mm
- Certification score: 88-93
- Calibration + Dimensional tests

**Prusa Mini+:**
- Default tier: Silver
- Expected tolerance: ±0.08-0.10mm
- Certification score: 85-90
- Calibration test only

## Job Submission Guide

### For Requesters

1. **Prepare 3D model**
   - Supported formats: STL, OBJ
   - Ensure model is manifold (watertight)
   - Check for printability issues
   - Recommended: Test print locally first

2. **Submit job**
   - Upload model file
   - Fill in requirements:
     - Material (PLA, PETG, ABS, etc.)
     - Quality tier needed (Silver/Gold/Platinum)
     - Required tolerance (if specific)
     - Quantity
     - Deadline (if time-sensitive)
     - Delivery location

3. **Automatic partitioning**
   - System analyzes model
   - Splits into printable parts if needed
   - Shows partition preview
   - Confirm or request manual adjustment

4. **Pricing & approval**
   - Review estimated cost
   - See assigned printers
   - Approve and pay deposit (50%)

5. **Track progress**
   - Real-time status updates
   - View which parts are printing
   - Get notifications at milestones

6. **QA review**
   - Review submitted photos
   - Approve or request corrections
   - Final payment on approval

7. **Delivery**
   - Parts shipped separately or assembled
   - Rate printer operators
   - Provide feedback

## QA Reviewer Guide

### Manual Review Process

1. **Access pending QA**
   - Dashboard → "QA Review Queue"
   - Sorted by submission time

2. **Review checklist**
   ```
   □ All required photos present (4+ angles)
   □ Photos are clear and well-lit
   □ Dimensional measurements provided
   □ Measurements within tolerance
   □ No visible defects
   □ AI analysis reviewed
   □ Part matches original model
   ```

3. **Decision making**

   **Approve if:**
   - QA score ≥ 80
   - Dimensions within tolerance
   - No critical defects
   - Photos are adequate

   **Request better photos if:**
   - Images blurry/dark
   - Missing required angles
   - Cannot assess quality

   **Reject & request reprint if:**
   - Visible defects (warping, layer separation)
   - Dimensions out of tolerance
   - Wrong material/color
   - Damaged part

4. **Provide feedback**
   - Always leave detailed notes
   - Be specific about issues
   - Suggest improvements

## Monitoring & Maintenance

### Health Checks

**Service health:**
```bash
# Check all services
docker-compose ps

# Individual service logs
docker-compose logs -f api-gateway
docker-compose logs -f job-orchestrator
```

**Database health:**
```bash
# Connect to database
docker-compose exec postgres psql -U swarmforge -d swarmforge

# Check active connections
SELECT count(*) FROM pg_stat_activity;

# Check table sizes
SELECT 
  tablename, 
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename))
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### Backup Strategy

**Database backups:**
```bash
# Manual backup
docker-compose exec postgres pg_dump -U swarmforge swarmforge > backup_$(date +%Y%m%d).sql

# Automated daily backups (add to crontab)
0 2 * * * docker-compose exec postgres pg_dump -U swarmforge swarmforge > /backups/swarmforge_$(date +\%Y\%m\%d).sql
```

**File storage backups:**
```bash
# Backup MinIO data
docker-compose exec minio mc mirror local/swarmforge-models /backups/models
```

### Performance Monitoring

**Prometheus metrics:**
- API request latency
- Database query performance
- Job processing time
- QA approval rates
- Printer utilization

**Key metrics to watch:**
- Average job completion time
- QA first-pass rate
- Printer success rate
- User satisfaction scores

## Troubleshooting

### Common Issues

**Services won't start:**
```bash
# Check Docker resources
docker system df

# Clean up
docker system prune -a

# Rebuild
docker-compose build --no-cache
docker-compose up -d
```

**Database connection errors:**
```bash
# Check if database is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Restart database
docker-compose restart postgres
```

**Frontend can't connect to API:**
- Check NEXT_PUBLIC_API_URL in .env
- Verify API Gateway is running: `curl http://localhost:8080/health`
- Check browser console for CORS errors

**QA photo uploads failing:**
- Check MinIO is running
- Verify bucket exists
- Check MinIO credentials in .env

## Support & Resources

- Documentation: `/docs`
- API Reference: http://localhost:8080/api/docs
- Community Forum: (link)
- Discord: (link)
- GitHub Issues: (link)

## License

MIT License - see LICENSE file
