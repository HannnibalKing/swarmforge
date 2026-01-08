import express from 'express';
import multer from 'multer';
import { S3Client, PutObjectCommand } from '@aws-sdk/client-s3';
import sharp from 'sharp';
import axios from 'axios';

const app = express();
const upload = multer({ storage: multer.memoryStorage() });

// MinIO/S3 client
const s3Client = new S3Client({
  endpoint: process.env.MINIO_ENDPOINT || 'http://localhost:9000',
  region: 'us-east-1',
  credentials: {
    accessKeyId: process.env.MINIO_ACCESS_KEY || 'minioadmin',
    secretAccessKey: process.env.MINIO_SECRET_KEY || 'minioadmin',
  },
  forcePathStyle: true,
});

interface QASubmission {
  jobPartId: string;
  printerId: string;
  submitterId: string;
  photos: string[];
  dimensions: {
    x: number;
    y: number;
    z: number;
  };
  weight?: number;
  measurementMethod: string;
  surfaceQualityScore: number;
  notes?: string;
}

interface AIAnalysisResult {
  defects: string[];
  confidence: number;
  overallQuality: number;
  flagged: boolean;
  details: {
    layerLines: number;
    warping: number;
    stringing: number;
    surfaceDefects: number;
  };
}

/**
 * AI-powered defect detection
 * This would integrate with a computer vision model (e.g., TensorFlow, PyTorch)
 */
async function analyzePhotos(photoUrls: string[]): Promise<AIAnalysisResult> {
  try {
    // In production, this would call an AI model endpoint
    // For now, placeholder logic
    const aiEndpoint = process.env.AI_MODEL_ENDPOINT;
    
    if (aiEndpoint) {
      const response = await axios.post(aiEndpoint, {
        images: photoUrls,
        model: 'defect-detection-v1',
      });
      
      return response.data;
    }
    
    // Fallback: rule-based analysis
    return performRuleBasedAnalysis(photoUrls);
  } catch (error) {
    console.error('AI analysis failed:', error);
    return {
      defects: [],
      confidence: 0,
      overallQuality: 70,
      flagged: false,
      details: {
        layerLines: 0,
        warping: 0,
        stringing: 0,
        surfaceDefects: 0,
      },
    };
  }
}

function performRuleBasedAnalysis(photoUrls: string[]): AIAnalysisResult {
  // Simple heuristics (would be replaced with actual CV model)
  const defects: string[] = [];
  let overallQuality = 90;
  
  // Placeholder: In reality, analyze image data
  // Check for common issues:
  // - Layer separation (visible gaps)
  // - Warping (curved base)
  // - Stringing (thin filament threads)
  // - Surface defects (blobs, zits)
  // - Under/over extrusion
  
  const flagged = overallQuality < 70;
  
  return {
    defects,
    confidence: 60, // Lower confidence for rule-based
    overallQuality,
    flagged,
    details: {
      layerLines: 5,
      warping: 0,
      stringing: 2,
      surfaceDefects: 1,
    },
  };
}

/**
 * Validate dimensional accuracy
 */
function validateDimensions(
  expected: { x: number; y: number; z: number },
  actual: { x: number; y: number; z: number },
  tolerance: number
): { passed: boolean; errors: { [key: string]: number } } {
  const errors: { [key: string]: number } = {};
  let passed = true;
  
  for (const dim of ['x', 'y', 'z']) {
    const error = Math.abs(expected[dim as keyof typeof expected] - actual[dim as keyof typeof actual]);
    if (error > tolerance) {
      errors[dim] = error;
      passed = false;
    }
  }
  
  return { passed, errors };
}

/**
 * Calculate QA score
 */
function calculateQAScore(
  aiAnalysis: AIAnalysisResult,
  dimensionalCheck: { passed: boolean; errors: { [key: string]: number } },
  surfaceQualityScore: number
): number {
  let score = 100;
  
  // Deduct for AI-detected defects
  score -= aiAnalysis.details.layerLines * 2;
  score -= aiAnalysis.details.warping * 5;
  score -= aiAnalysis.details.stringing * 1;
  score -= aiAnalysis.details.surfaceDefects * 3;
  
  // Deduct for dimensional errors
  if (!dimensionalCheck.passed) {
    const avgError = Object.values(dimensionalCheck.errors).reduce((a, b) => a + b, 0) / 
                     Object.keys(dimensionalCheck.errors).length;
    score -= avgError * 10; // -10 points per 0.1mm error
  }
  
  // Factor in manual surface quality score
  score = (score * 0.7) + (surfaceQualityScore * 0.3);
  
  return Math.max(0, Math.min(100, score));
}

// Routes

app.use(express.json());

app.post('/api/qa/submit', upload.array('photos', 10), async (req, res) => {
  try {
    const files = req.files as Express.Multer.File[];
    const submission: QASubmission = JSON.parse(req.body.data);
    
    // Upload photos to S3/MinIO
    const photoUrls: string[] = [];
    for (const file of files) {
      const key = `qa-photos/${submission.jobPartId}/${Date.now()}-${file.originalname}`;
      
      // Optimize image
      const optimized = await sharp(file.buffer)
        .resize(1920, 1920, { fit: 'inside', withoutEnlargement: true })
        .jpeg({ quality: 85 })
        .toBuffer();
      
      await s3Client.send(new PutObjectCommand({
        Bucket: 'swarmforge-qa-photos',
        Key: key,
        Body: optimized,
        ContentType: 'image/jpeg',
      }));
      
      const url = `${process.env.MINIO_ENDPOINT}/swarmforge-qa-photos/${key}`;
      photoUrls.push(url);
    }
    
    // Perform AI analysis
    const aiAnalysis = await analyzePhotos(photoUrls);
    
    // TODO: Fetch expected dimensions from database
    const expectedDimensions = { x: 20, y: 20, z: 20 }; // Placeholder
    const dimensionalCheck = validateDimensions(
      expectedDimensions,
      submission.dimensions,
      0.1 // tolerance in mm
    );
    
    // Calculate overall QA score
    const qaScore = calculateQAScore(
      aiAnalysis,
      dimensionalCheck,
      submission.surfaceQualityScore
    );
    
    // Determine if manual review is needed
    const requiresReview = aiAnalysis.flagged || !dimensionalCheck.passed || qaScore < 80;
    
    // TODO: Save to database
    const qaSubmissionRecord = {
      ...submission,
      photoUrls,
      aiAnalysis,
      dimensionalCheck,
      qaScore,
      status: requiresReview ? 'pending_review' : 'approved',
      submittedAt: new Date(),
    };
    
    console.log('QA Submission:', qaSubmissionRecord);
    
    res.json({
      success: true,
      qaScore,
      requiresReview,
      aiAnalysis,
      dimensionalCheck,
    });
  } catch (error) {
    console.error('QA submission error:', error);
    res.status(500).json({ error: 'Failed to process QA submission' });
  }
});

app.get('/api/qa/pending', async (req, res) => {
  // TODO: Fetch from database
  res.json({
    pending: [],
  });
});

app.post('/api/qa/review/:id', async (req, res) => {
  const { id } = req.params;
  const { approved, notes } = req.body;
  
  // TODO: Update database
  // TODO: If approved, update part status and trigger payment
  // TODO: If rejected, create reprint request
  
  res.json({
    success: true,
    id,
    approved,
  });
});

app.get('/health', (req, res) => {
  res.json({ status: 'healthy' });
});

const PORT = process.env.QA_SERVICE_PORT || 3001;
app.listen(PORT, () => {
  console.log(`QA Service running on port ${PORT}`);
});
