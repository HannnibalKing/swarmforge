# SwarmForge Quality Assurance System

## Overview
The QA system ensures that all printed parts meet specified quality standards before assembly and delivery. It combines automated AI analysis with manual review processes.

## QA Workflow

### 1. Pre-Print QA
Before printing begins, each part goes through automated validation:

```
Model Upload → STL Validation → Printability Check → Slicing Verification
```

**Checks:**
- **Manifold verification**: Ensure 3D model is watertight (no holes)
- **Printability analysis**: Detect overhangs, thin walls, minimum feature size
- **Material compatibility**: Verify selected material works for part geometry
- **Print time estimation**: Calculate realistic time based on printer capabilities

### 2. During Print QA
For printers with cameras (Bambu X1C, P1S):

- **First layer check**: AI analyzes first layer adhesion (5 minutes in)
- **Mid-print monitoring**: Check for warping, layer shifts (at 50%)
- **Spaghetti detection**: Identify print failures early

### 3. Post-Print QA

#### Phase 1: Photo Submission (Required)
Printer operators must submit:
- **4+ photos from different angles**
  - Front view
  - Back view
  - Top view
  - Close-up of critical features
- **Lighting requirements**: Well-lit, neutral background
- **Metadata**: Timestamp, printer ID, part ID

#### Phase 2: AI Analysis
Automated defect detection using computer vision:

```python
# Defect Detection Model
defects_checked = [
    'layer_separation',    # Gaps between layers
    'warping',             # Curled edges/corners
    'stringing',           # Thin filament threads
    'under_extrusion',     # Gaps in solid areas
    'over_extrusion',      # Bulging, blobs
    'surface_defects',     # Zits, blobs, scratches
    'support_marks',       # Poorly removed supports
    'dimensional_issues'   # Visible size discrepancies
]
```

**AI Scoring:**
- Each defect type assigned severity (1-10)
- Confidence score for each detection
- Overall quality score (0-100)
- Automatic flagging if score < 70

#### Phase 3: Dimensional Verification
Printer operators measure parts using calipers/micrometers:

```typescript
interface DimensionalCheck {
  expected: { x: number, y: number, z: number };
  actual: { x: number, y: number, z: number };
  tolerance: number; // mm
  measurement_tool: 'caliper' | 'micrometer' | '3d_scanner';
  critical_dimensions?: { [feature: string]: number };
}
```

**Tolerance Levels by Tier:**
- **Platinum**: ±0.05mm
- **Gold**: ±0.08mm  
- **Silver**: ±0.10mm

**Measurement Protocol:**
1. Wait for part to cool completely (≥30 minutes)
2. Measure at room temperature (20-25°C)
3. Take 3 measurements per dimension, use average
4. Document measurement locations

#### Phase 4: Manual Review
If AI flags or dimensions fail:

**QA Reviewer Checklist:**
- [ ] Review all submitted photos
- [ ] Verify dimensional measurements make sense
- [ ] Check for defects AI might have missed
- [ ] Assess overall fit for purpose
- [ ] Decision: Approve / Request Better Photos / Reject / Reprint

**Review Criteria:**
```javascript
const qaDecision = (submission) => {
  if (submission.qaScore >= 90 && !submission.aiFlagged) {
    return 'auto_approve';
  }
  
  if (submission.qaScore >= 80 && submission.dimensionalCheck.passed) {
    return 'approve_with_notes';
  }
  
  if (submission.qaScore >= 70) {
    return 'manual_review_required';
  }
  
  if (submission.criticalDefects.length > 0) {
    return 'reject_reprint';
  }
  
  return 'request_better_photos';
};
```

### 4. Final Assembly QA

After all parts are manufactured:

**Fit Testing:**
- All parts must assemble without force
- Tolerances verified at connection points
- Moving parts tested for proper clearance
- Structural integrity verified

**Assembly Checklist:**
```
□ All parts present and accounted for
□ Parts fit together as designed
□ No visible gaps at joints
□ Fasteners (if any) fit properly
□ Overall dimensions match CAD
□ Surface finish consistent across parts
□ No color variations (same material batch preferred)
```

## Calibration Test Prints

### For Printer Certification

#### Test 1: Calibration Cube (Required for All Tiers)
- **Model**: 20mm × 20mm × 20mm cube
- **Material**: PLA
- **Settings**: 0.2mm layer height, 20% infill
- **Measurements**: X, Y, Z dimensions
- **Pass criteria**: 
  - Platinum: 19.95-20.05mm all dimensions
  - Gold: 19.92-20.08mm
  - Silver: 19.90-20.10mm

#### Test 2: Dimensional Accuracy Test (Gold+)
- **Model**: Tolerance test piece with various features
- **Features**: Holes (5mm, 10mm), slots, bosses
- **Pass criteria**: All features within tier tolerance

#### Test 3: Bridging Test (Platinum only)
- **Model**: Bridge test with gaps (10mm, 20mm, 30mm, 40mm, 50mm)
- **Pass criteria**: Clean bridges up to 40mm with <0.5mm sag

#### Test 4: Overhang Test (Platinum only)
- **Model**: Overhangs at 30°, 45°, 60°, 70°
- **Pass criteria**: Clean surfaces up to 60° without supports

## Reputation System

### Printer Reputation Score
```typescript
interface PrinterReputation {
  total_jobs: number;
  successful_jobs: number;
  success_rate: number; // percentage
  average_qa_score: number; // 0-100
  on_time_delivery_rate: number; // percentage
  qa_first_pass_rate: number; // approved without resubmission
  user_rating: number; // 1-5 stars
  tier_compliance: boolean; // maintaining tier standards
}
```

**Reputation Impact:**
- **Success rate < 80%**: Warning issued
- **Success rate < 70%**: Tier downgrade review
- **Success rate < 60%**: Printer suspended
- **QA score consistently < 85**: Recertification required
- **Multiple user complaints**: Manual review

### User (Requester) Reputation
- Payment reliability
- Communication quality
- Realistic expectations
- Review fairness

## AI Defect Detection Implementation

### Model Architecture
```python
# Pseudo-code for defect detection
import tensorflow as tf
from tensorflow.keras import layers

def build_defect_detector():
    model = tf.keras.Sequential([
        # Image preprocessing
        layers.Rescaling(1./255),
        
        # Feature extraction (CNN)
        layers.Conv2D(32, 3, activation='relu'),
        layers.MaxPooling2D(),
        layers.Conv2D(64, 3, activation='relu'),
        layers.MaxPooling2D(),
        layers.Conv2D(128, 3, activation='relu'),
        layers.MaxPooling2D(),
        
        # Classification
        layers.Flatten(),
        layers.Dense(256, activation='relu'),
        layers.Dropout(0.5),
        
        # Multi-label output
        layers.Dense(len(DEFECT_TYPES), activation='sigmoid')
    ])
    
    return model

# Training data requirements:
# - 10,000+ labeled images of 3D prints
# - Multiple angles per print
# - Various defect types and severities
# - Balanced dataset across materials and printers
```

### Training Data Collection
- Community-sourced labeled images
- Each print requires:
  - 4+ photos
  - Defect labels (if any)
  - Severity ratings
  - Print parameters (material, layer height, etc.)

## Dispute Resolution

### When QA is Disputed

**Requester disputes rejection:**
1. Submit additional photos/evidence
2. Request third-party review
3. Escalate to admin if disagreement persists

**Printer disputes rejection:**
1. Submit better quality photos
2. Provide detailed measurements
3. Request re-review by different QA reviewer

**Resolution process:**
- Independent QA reviewer assigned
- Both parties submit evidence
- Final decision by admin
- Compensation determined based on fault

## Quality Metrics Dashboard

### For Printers
```
┌─────────────────────────────────────┐
│ Your QA Performance                 │
├─────────────────────────────────────┤
│ First Pass Rate:        94%  ✓      │
│ Average QA Score:       92/100      │
│ Rejection Rate:         3%          │
│ Resubmission Rate:      3%          │
│ Current Tier:           Gold        │
│ Next Cert Review:       45 days     │
└─────────────────────────────────────┘
```

### For Requesters
```
┌─────────────────────────────────────┐
│ Job #12345 QA Status                │
├─────────────────────────────────────┤
│ Parts Approved:         8/12        │
│ Parts In Review:        3/12        │
│ Parts Rejected:         1/12        │
│ Average Part Score:     88/100      │
│ Estimated Completion:   2 days      │
└─────────────────────────────────────┘
```

## Best Practices

### For Printer Operators
1. **Calibrate regularly** - Run test prints monthly
2. **Clean photos** - Good lighting, neutral background
3. **Accurate measurements** - Use proper tools, cool parts first
4. **Document issues** - Report problems early
5. **Learn from rejections** - Review feedback, improve settings

### For Requesters
1. **Clear requirements** - Specify tolerance needs upfront
2. **Realistic tolerances** - Don't over-specify precision
3. **Material selection** - Choose appropriate for application
4. **Fair reviews** - Rate based on agreed requirements
5. **Communication** - Respond to questions promptly

## Future Enhancements

1. **Automated dimensional verification** using uploaded measurement photos
2. **Real-time print monitoring** via webcam streams
3. **Material certification** tracking batch numbers for consistency
4. **3D scanner integration** for complex geometries
5. **Blockchain verification** for immutable QA records
6. **Machine learning improvements** as more data is collected
