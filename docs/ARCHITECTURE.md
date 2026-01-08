# SwarmForge Architecture Documentation

## System Overview

SwarmForge is a distributed 3D printing platform built on microservices architecture, designed to handle job distribution, printer certification, quality assurance, and real-time communication at scale.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         Frontend (Next.js)                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │Dashboard │  │Jobs List │  │Printers  │  │QA Review │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└───────────┬─────────────────────────────────────┬───────────┘
            │ REST API                    WebSocket│
            │                                      │
┌───────────▼──────────────────────────────────────▼───────────┐
│                    API Gateway (Go)                          │
│              Authentication, Rate Limiting                    │
└───────────┬──────────────────────────────────────────────────┘
            │
┌───────────▼──────────────────────────────────────────────────┐
│                     Service Layer                            │
│                                                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Job       │  │  Printer    │  │    QA       │         │
│  │Orchestrator │  │  Service    │  │  Service    │         │
│  │    (Go)     │  │    (Go)     │  │ (Node.js)   │         │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘         │
│         │                 │                 │                 │
│  ┌──────▼──────┐  ┌──────▼──────┐  ┌──────▼──────┐         │
│  │   Model     │  │  WebSocket  │  │  Payment    │         │
│  │ Processor   │  │   Server    │  │  Service    │         │
│  │    (Go)     │  │ (Node.js)   │  │   (Go)      │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└───────────┬──────────────────────────────────────────────────┘
            │
┌───────────▼──────────────────────────────────────────────────┐
│                     Data Layer                               │
│                                                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ PostgreSQL  │  │    Redis    │  │  RabbitMQ   │         │
│  │             │  │             │  │             │         │
│  │ Relational  │  │   Cache &   │  │   Message   │         │
│  │    Data     │  │   Session   │  │    Queue    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                               │
│  ┌─────────────┐                                            │
│  │   MinIO     │                                            │
│  │             │                                            │
│  │  S3 Object  │                                            │
│  │  Storage    │                                            │
│  └─────────────┘                                            │
└───────────────────────────────────────────────────────────────┘
```

## Service Responsibilities

### 1. API Gateway (Go)
**Purpose**: Single entry point for all client requests

**Responsibilities:**
- Request routing to appropriate services
- JWT authentication and authorization
- Rate limiting and throttling
- Request/response logging
- CORS handling
- API versioning

**Key Endpoints:**
```
POST   /api/v1/auth/login
POST   /api/v1/auth/register
GET    /api/v1/users/me

POST   /api/v1/jobs
GET    /api/v1/jobs/:id
GET    /api/v1/jobs/:id/parts
PATCH  /api/v1/jobs/:id/status

POST   /api/v1/printers
GET    /api/v1/printers/:id
GET    /api/v1/printers/available
POST   /api/v1/printers/:id/certify

POST   /api/v1/qa/submit
GET    /api/v1/qa/pending
POST   /api/v1/qa/review/:id
```

**Technologies:**
- Gin framework
- JWT for authentication
- Redis for session storage

### 2. Job Orchestrator (Go)
**Purpose**: Manages job lifecycle and part distribution

**Responsibilities:**
- Process incoming print jobs
- Partition 3D models into printable parts
- Match parts with suitable printers
- Monitor job progress
- Handle timeouts and failures
- Reassign failed parts

**Algorithms:**

#### Part Assignment Algorithm
```go
func AssignPartToPrinter(part JobPart, requirements Requirements) *Printer {
    // 1. Filter eligible printers
    eligible := FilterPrinters(
        ByPrintVolume(part.Dimensions),
        ByMaterial(requirements.Material),
        ByTier(requirements.MinTier),
        ByTolerance(requirements.Tolerance),
        ByAvailability(true),
    )
    
    // 2. Score each printer
    scores := []PrinterScore{}
    for _, printer := range eligible {
        score := CalculatePrinterScore(printer, part, requirements)
        scores = append(scores, PrinterScore{printer, score})
    }
    
    // 3. Sort by score (descending)
    sort.Slice(scores, func(i, j int) bool {
        return scores[i].Score > scores[j].Score
    })
    
    // 4. Return best match
    return scores[0].Printer
}

func CalculatePrinterScore(printer *Printer, part JobPart, req Requirements) float64 {
    score := 0.0
    
    // Location proximity (30%)
    distance := CalculateDistance(req.DeliveryLocation, printer.Location)
    proximityScore := 1.0 / (1.0 + distance/1000.0) // Closer = better
    score += proximityScore * 30.0
    
    // Reputation (30%)
    score += (printer.SuccessRate / 100.0) * 30.0
    
    // Certification tier (25%)
    tierScore := map[string]float64{
        "platinum": 1.0,
        "gold":     0.8,
        "silver":   0.6,
    }[printer.Tier]
    score += tierScore * 25.0
    
    // Availability/turnaround (15%)
    if printer.CurrentJobs == 0 {
        score += 15.0
    } else {
        score += 15.0 / float64(printer.CurrentJobs+1)
    }
    
    return score
}
```

**Job State Machine:**
```
pending → analyzing → partitioned → assigned → printing 
    → qa_review → assembly → completed
                  ↓
              (failed/cancelled)
```

### 3. Printer Service (Go)
**Purpose**: Manages printer registration, certification, and capabilities

**Responsibilities:**
- Printer registration and profile management
- Certification test submission and evaluation
- Capability tracking (materials, volumes, etc.)
- Printer grouping by tier
- Calibration score calculation
- Reputation management

**Printer Profiles:**
```go
type PrinterProfile struct {
    ID               string
    Manufacturer     string  // "Bambu Lab", "Prusa Research"
    Model            string  // "X1C", "MK4"
    Tier             string  // "platinum", "gold", "silver"
    CalibrationScore int     // 0-100
    
    // Capabilities
    MaxVolume        [3]int      // [x, y, z] in mm
    Materials        []string    // ["PLA", "PETG", "ABS"]
    NozzleDiameter   float64     // mm
    LayerHeightRange [2]float64  // [min, max] mm
    
    // Performance
    ExpectedTolerance float64    // mm
    SuccessRate       float64    // 0-100%
    AvgQAScore        float64    // 0-100
    
    // Location
    Location         GeoPoint
    Availability     bool
}
```

**Certification Scoring:**
- Dimensional accuracy: 50% weight
- Surface quality: 30% weight
- Consistency: 20% weight

### 4. QA Service (Node.js/TypeScript)
**Purpose**: Quality assurance and defect detection

**Responsibilities:**
- Photo upload and storage
- AI-powered defect detection
- Dimensional verification
- Manual review workflow
- QA scoring and approval

**AI Integration:**
```typescript
// Defect detection pipeline
async function analyzeQA(submission: QASubmission): Promise<QAResult> {
    // 1. Upload photos to S3
    const photoUrls = await uploadPhotos(submission.photos);
    
    // 2. Run AI defect detection
    const aiAnalysis = await detectDefects(photoUrls);
    
    // 3. Verify dimensions
    const dimensionalCheck = verifyDimensions(
        submission.expected,
        submission.actual,
        submission.tolerance
    );
    
    // 4. Calculate overall score
    const qaScore = calculateScore(aiAnalysis, dimensionalCheck);
    
    // 5. Determine if manual review needed
    const requiresReview = qaScore < 80 || aiAnalysis.flagged;
    
    return {
        score: qaScore,
        requiresReview,
        aiAnalysis,
        dimensionalCheck
    };
}
```

### 5. Model Processor (Go)
**Purpose**: 3D model analysis and partitioning

**Responsibilities:**
- STL file validation
- Model slicing and partitioning
- Orientation optimization
- Support generation assessment
- Print time estimation
- Material usage calculation

**Partitioning Strategy:**
```go
func PartitionModel(model STLFile, maxVolume [3]int) []ModelPart {
    // 1. Analyze model bounding box
    bounds := model.CalculateBounds()
    
    // 2. If fits in single printer, no partition needed
    if FitsInVolume(bounds, maxVolume) {
        return []ModelPart{{File: model, PartNumber: 1}}
    }
    
    // 3. Find optimal split planes
    splitPlanes := FindOptimalSplits(model, maxVolume)
    
    // 4. Split model along planes
    parts := SplitModel(model, splitPlanes)
    
    // 5. Add connector features (dovetails, pins, etc.)
    for i := range parts {
        parts[i] = AddConnectors(parts[i])
    }
    
    return parts
}
```

### 6. WebSocket Server (Node.js)
**Purpose**: Real-time updates to clients

**Responsibilities:**
- WebSocket connection management
- Pub/sub integration with Redis
- Real-time job status updates
- Printer status broadcasting
- QA notification delivery

**Message Types:**
```typescript
type WSMessage = 
    | { type: 'job_update', jobId: string, status: string }
    | { type: 'part_update', partId: string, status: string }
    | { type: 'printer_status', printerId: string, available: boolean }
    | { type: 'qa_submitted', partId: string }
    | { type: 'qa_approved', partId: string }
    | { type: 'notification', message: string };
```

## Data Flow Examples

### Job Submission Flow
```
1. User uploads STL → Frontend
2. Frontend → API Gateway → Job Orchestrator
3. Job Orchestrator:
   - Save job to PostgreSQL
   - Upload STL to MinIO
   - Queue for analysis (RabbitMQ)
4. Model Processor:
   - Analyze STL
   - Partition if needed
   - Generate G-code
5. Job Orchestrator:
   - Create job parts
   - Assign to printers (algorithm)
   - Publish updates (Redis)
6. WebSocket → Notify user of assignment
```

### QA Submission Flow
```
1. Printer uploads photos → QA Service
2. QA Service:
   - Store photos in MinIO
   - Run AI defect detection
   - Validate dimensions
   - Calculate QA score
3. If score < 80 or flagged:
   - Create review task
   - Notify QA reviewer
4. Else:
   - Auto-approve
   - Update part status
   - Trigger payment
5. WebSocket → Notify requester
```

## Scalability Considerations

### Horizontal Scaling
- **Stateless services**: All services can scale horizontally
- **Load balancing**: Nginx/HAProxy for API Gateway
- **Database read replicas**: PostgreSQL replication
- **Caching layer**: Redis for frequently accessed data

### Performance Optimization
- **CDN**: Serve static assets and STL files
- **Image optimization**: Sharp for photo compression
- **Query optimization**: Indexed database queries
- **Background jobs**: RabbitMQ for async processing
- **Connection pooling**: Database connection management

### Monitoring & Observability
- **Metrics**: Prometheus + Grafana
- **Logging**: ELK stack (Elasticsearch, Logstash, Kibana)
- **Tracing**: Jaeger for distributed tracing
- **Alerting**: PagerDuty for critical issues

## Security

### Authentication & Authorization
- JWT tokens with refresh mechanism
- Role-based access control (RBAC)
- API key authentication for printers

### Data Protection
- TLS/SSL for all communications
- Encrypted storage for sensitive data
- Signed URLs for S3 access
- Rate limiting to prevent abuse

### Compliance
- GDPR compliance for user data
- Data retention policies
- Audit logging for all transactions

## Deployment

### Development
```bash
docker-compose up
```

### Production (Kubernetes)
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    spec:
      containers:
      - name: api-gateway
        image: swarmforge/api-gateway:latest
        resources:
          limits:
            cpu: "1"
            memory: "512Mi"
```

## Future Architecture Enhancements

1. **Event Sourcing**: Complete audit trail of all state changes
2. **CQRS**: Separate read and write models for better performance
3. **GraphQL**: Flexible API querying
4. **Service Mesh**: Istio for advanced traffic management
5. **Multi-region**: Global distribution for lower latency
