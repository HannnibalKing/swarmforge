-- SwarmForge Database Schema

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('requester', 'printer', 'qa_reviewer', 'admin')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT true
);

-- Printer profiles
CREATE TABLE printers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    manufacturer VARCHAR(100) NOT NULL, -- Bambu, Prusa, etc.
    model VARCHAR(100) NOT NULL, -- X1C, P1S, MK4, etc.
    tier VARCHAR(20) NOT NULL CHECK (tier IN ('silver', 'gold', 'platinum')),
    calibration_score INTEGER DEFAULT 0 CHECK (calibration_score >= 0 AND calibration_score <= 100),
    certification_status VARCHAR(50) DEFAULT 'pending' CHECK (certification_status IN ('pending', 'certified', 'suspended', 'revoked')),
    certification_date TIMESTAMP,
    
    -- Capabilities
    max_print_volume_x INTEGER NOT NULL, -- mm
    max_print_volume_y INTEGER NOT NULL,
    max_print_volume_z INTEGER NOT NULL,
    supported_materials TEXT[], -- ['PLA', 'PETG', 'ABS', 'TPU']
    nozzle_diameter DECIMAL(4,2) DEFAULT 0.4, -- mm
    layer_height_min DECIMAL(4,3) DEFAULT 0.1, -- mm
    layer_height_max DECIMAL(4,3) DEFAULT 0.3,
    
    -- Status
    is_available BOOLEAN DEFAULT true,
    location_lat DECIMAL(10, 8),
    location_lng DECIMAL(11, 8),
    location_city VARCHAR(100),
    location_country VARCHAR(100),
    
    -- Statistics
    total_jobs_completed INTEGER DEFAULT 0,
    success_rate DECIMAL(5,2) DEFAULT 100.00,
    average_qa_score DECIMAL(5,2) DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_printers_user ON printers(user_id);
CREATE INDEX idx_printers_tier ON printers(tier);
CREATE INDEX idx_printers_available ON printers(is_available);
CREATE INDEX idx_printers_location ON printers(location_lat, location_lng);

-- Certification tests
CREATE TABLE certification_tests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    printer_id UUID REFERENCES printers(id) ON DELETE CASCADE,
    test_type VARCHAR(100) NOT NULL, -- 'calibration_cube', 'dimensional_accuracy', 'bridging', etc.
    test_file_url VARCHAR(500),
    photo_urls TEXT[], -- Array of QA photo URLs
    
    -- Measurements
    expected_dimensions JSONB, -- {"x": 20.0, "y": 20.0, "z": 20.0}
    actual_dimensions JSONB,
    tolerance_achieved DECIMAL(5,3), -- mm
    
    score INTEGER CHECK (score >= 0 AND score <= 100),
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'passed', 'failed')),
    reviewer_id UUID REFERENCES users(id),
    review_notes TEXT,
    
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reviewed_at TIMESTAMP
);

CREATE INDEX idx_cert_tests_printer ON certification_tests(printer_id);

-- Print jobs
CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_id UUID REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Model information
    model_file_url VARCHAR(500) NOT NULL,
    model_format VARCHAR(20) DEFAULT 'STL',
    total_parts INTEGER DEFAULT 1,
    
    -- Requirements
    required_material VARCHAR(50) NOT NULL,
    required_quality_tier VARCHAR(20) CHECK (required_quality_tier IN ('silver', 'gold', 'platinum')),
    required_tolerance DECIMAL(5,3), -- mm
    deadline TIMESTAMP,
    
    -- Job status
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN (
        'pending', 'analyzing', 'partitioned', 'assigned', 
        'printing', 'qa_review', 'assembly', 'completed', 
        'cancelled', 'failed'
    )),
    
    -- Pricing
    estimated_cost DECIMAL(10,2),
    final_cost DECIMAL(10,2),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX idx_jobs_requester ON jobs(requester_id);
CREATE INDEX idx_jobs_status ON jobs(status);

-- Job parts (individual pieces of a split model)
CREATE TABLE job_parts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID REFERENCES jobs(id) ON DELETE CASCADE,
    part_number INTEGER NOT NULL,
    part_name VARCHAR(255),
    
    -- Part file
    stl_file_url VARCHAR(500) NOT NULL,
    gcode_file_url VARCHAR(500),
    
    -- Assignment
    assigned_printer_id UUID REFERENCES printers(id),
    assigned_at TIMESTAMP,
    
    -- Print details
    estimated_print_time INTEGER, -- minutes
    estimated_material_weight DECIMAL(8,2), -- grams
    layer_height DECIMAL(4,3),
    infill_percentage INTEGER,
    supports_required BOOLEAN DEFAULT false,
    
    -- Status
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN (
        'pending', 'assigned', 'printing', 'printed', 
        'qa_pending', 'qa_approved', 'qa_rejected', 'shipped'
    )),
    
    -- Redundancy (for critical parts, assign to multiple printers)
    is_redundant BOOLEAN DEFAULT false,
    primary_part_id UUID REFERENCES job_parts(id),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(job_id, part_number)
);

CREATE INDEX idx_job_parts_job ON job_parts(job_id);
CREATE INDEX idx_job_parts_printer ON job_parts(assigned_printer_id);
CREATE INDEX idx_job_parts_status ON job_parts(status);

-- QA submissions
CREATE TABLE qa_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_part_id UUID REFERENCES job_parts(id) ON DELETE CASCADE,
    printer_id UUID REFERENCES printers(id),
    submitter_id UUID REFERENCES users(id),
    
    -- Photos
    photo_urls TEXT[] NOT NULL, -- Multi-angle photos
    photo_metadata JSONB, -- Camera info, lighting, etc.
    
    -- Measurements
    dimensions JSONB, -- Actual measured dimensions
    weight_grams DECIMAL(8,2),
    measurement_method VARCHAR(100), -- 'caliper', 'micrometer', etc.
    
    -- Visual inspection
    surface_quality_score INTEGER CHECK (surface_quality_score >= 1 AND surface_quality_score <= 10),
    defects_detected TEXT[], -- ['warping', 'layer_separation', 'stringing']
    notes TEXT,
    
    -- AI analysis
    ai_analysis_result JSONB,
    ai_defect_confidence DECIMAL(5,2), -- 0-100%
    ai_flagged BOOLEAN DEFAULT false,
    
    -- Review
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'requires_reprint')),
    reviewer_id UUID REFERENCES users(id),
    review_notes TEXT,
    reviewed_at TIMESTAMP,
    
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_qa_submissions_part ON qa_submissions(job_part_id);
CREATE INDEX idx_qa_submissions_status ON qa_submissions(status);
CREATE INDEX idx_qa_submissions_ai_flagged ON qa_submissions(ai_flagged);

-- Reputation and ratings
CREATE TABLE ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID REFERENCES jobs(id) ON DELETE CASCADE,
    printer_id UUID REFERENCES printers(id) ON DELETE CASCADE,
    rater_id UUID REFERENCES users(id),
    
    quality_score INTEGER CHECK (quality_score >= 1 AND quality_score <= 5),
    communication_score INTEGER CHECK (communication_score >= 1 AND communication_score <= 5),
    timeliness_score INTEGER CHECK (timeliness_score >= 1 AND timeliness_score <= 5),
    
    comment TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ratings_printer ON ratings(printer_id);
CREATE INDEX idx_ratings_job ON ratings(job_id);

-- Transactions/Payments
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID REFERENCES jobs(id),
    job_part_id UUID REFERENCES job_parts(id),
    
    from_user_id UUID REFERENCES users(id),
    to_user_id UUID REFERENCES users(id),
    
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    
    transaction_type VARCHAR(50) CHECK (transaction_type IN ('payment', 'refund', 'escrow_hold', 'escrow_release')),
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed', 'refunded')),
    
    payment_provider VARCHAR(100),
    external_transaction_id VARCHAR(255),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_job ON transactions(job_id);
CREATE INDEX idx_transactions_from_user ON transactions(from_user_id);
CREATE INDEX idx_transactions_to_user ON transactions(to_user_id);

-- Notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    
    type VARCHAR(100) NOT NULL, -- 'job_assigned', 'qa_approved', 'payment_received', etc.
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    
    related_job_id UUID REFERENCES jobs(id),
    related_part_id UUID REFERENCES job_parts(id),
    
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read);

-- Audit log
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100),
    entity_id UUID,
    changes JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);
