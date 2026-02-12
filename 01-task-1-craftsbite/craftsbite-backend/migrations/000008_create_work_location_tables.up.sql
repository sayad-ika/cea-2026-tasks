CREATE TABLE work_location_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    location VARCHAR(20) NOT NULL CHECK (location IN ('office', 'wfh')),
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    reason VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (user_id, date)
);

CREATE INDEX idx_work_location_statuses_user_date ON work_location_statuses(user_id, date);
CREATE INDEX idx_work_location_statuses_date ON work_location_statuses(date);

CREATE TABLE global_work_location_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    location VARCHAR(20) NOT NULL CHECK (location IN ('wfh')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    reason VARCHAR(255),
    declared_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CHECK (end_date >= start_date)
);

CREATE INDEX idx_global_work_location_active_range
    ON global_work_location_policies(is_active, start_date, end_date);

CREATE TABLE work_location_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    previous_location VARCHAR(20),
    new_location VARCHAR(20) NOT NULL,
    changed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    reason VARCHAR(255),
    action VARCHAR(30) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_work_location_history_user_date ON work_location_history(user_id, date);
CREATE INDEX idx_work_location_history_created_at ON work_location_history(created_at);
