-- =================================
-- Tajeor Blockchain Explorer Schema
-- =================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Drop tables if they exist (for development only)
-- DROP TABLE IF EXISTS transactions CASCADE;
-- DROP TABLE IF EXISTS blocks CASCADE;
-- DROP TABLE IF EXISTS validators CASCADE;
-- DROP TABLE IF EXISTS accounts CASCADE;
-- DROP TABLE IF EXISTS delegations CASCADE;
-- DROP TABLE IF EXISTS validator_history CASCADE;
-- DROP TABLE IF EXISTS network_stats CASCADE;

-- Accounts table
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    address VARCHAR(45) UNIQUE NOT NULL,
    balance DECIMAL(20, 6) DEFAULT 0,
    account_number INTEGER,
    sequence_number INTEGER DEFAULT 0,
    key_name VARCHAR(100),
    account_type VARCHAR(20) DEFAULT 'base',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Validators table
CREATE TABLE validators (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    moniker VARCHAR(100) NOT NULL,
    operator_address VARCHAR(52) UNIQUE NOT NULL,
    consensus_address VARCHAR(52),
    delegator_address VARCHAR(45),
    commission_rate DECIMAL(5, 4) NOT NULL,
    max_commission_rate DECIMAL(5, 4),
    max_commission_change_rate DECIMAL(5, 4),
    min_self_delegation DECIMAL(20, 6) NOT NULL,
    self_delegation DECIMAL(20, 6) DEFAULT 0,
    total_delegation DECIMAL(20, 6) DEFAULT 0,
    voting_power DECIMAL(15, 6) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    jailed BOOLEAN DEFAULT FALSE,
    created_by_key VARCHAR(100),
    website VARCHAR(255),
    description TEXT,
    identity VARCHAR(64),
    security_contact VARCHAR(255),
    details TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Blocks table
CREATE TABLE blocks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    height BIGINT UNIQUE NOT NULL,
    hash VARCHAR(64) UNIQUE NOT NULL,
    previous_hash VARCHAR(64),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    proposer_address VARCHAR(52),
    transaction_count INTEGER DEFAULT 0,
    size_bytes INTEGER,
    gas_limit BIGINT,
    gas_used BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Transactions table
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hash VARCHAR(64) UNIQUE NOT NULL,
    block_height BIGINT NOT NULL,
    block_hash VARCHAR(64),
    transaction_index INTEGER,
    from_address VARCHAR(45),
    to_address VARCHAR(45),
    amount DECIMAL(20, 6),
    fee DECIMAL(20, 6),
    gas_limit BIGINT,
    gas_used BIGINT,
    status VARCHAR(20) DEFAULT 'success',
    type VARCHAR(50),
    memo TEXT,
    raw_log TEXT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (block_height) REFERENCES blocks(height) ON DELETE CASCADE
);

-- Delegations table
CREATE TABLE delegations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    delegator_address VARCHAR(45) NOT NULL,
    validator_address VARCHAR(52) NOT NULL,
    amount DECIMAL(20, 6) NOT NULL,
    shares DECIMAL(20, 6) NOT NULL,
    height BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (delegator_address) REFERENCES accounts(address) ON DELETE CASCADE,
    FOREIGN KEY (validator_address) REFERENCES validators(operator_address) ON DELETE CASCADE
);

-- Validator history for tracking changes
CREATE TABLE validator_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    validator_id UUID NOT NULL,
    block_height BIGINT NOT NULL,
    voting_power DECIMAL(15, 6),
    commission_rate DECIMAL(5, 4),
    total_delegation DECIMAL(20, 6),
    status VARCHAR(20),
    jailed BOOLEAN,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (validator_id) REFERENCES validators(id) ON DELETE CASCADE
);

-- Network statistics
CREATE TABLE network_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    block_height BIGINT NOT NULL,
    total_validators INTEGER,
    active_validators INTEGER,
    total_supply DECIMAL(20, 6),
    bonded_tokens DECIMAL(20, 6),
    not_bonded_tokens DECIMAL(20, 6),
    staking_ratio DECIMAL(5, 4),
    inflation_rate DECIMAL(5, 4),
    community_pool DECIMAL(20, 6),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- API usage tracking
CREATE TABLE api_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code INTEGER,
    response_time_ms INTEGER,
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User sessions (if authentication is enabled)
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id VARCHAR(255) UNIQUE NOT NULL,
    user_id VARCHAR(100),
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    active BOOLEAN DEFAULT TRUE
);

-- Create indexes for performance
CREATE INDEX idx_accounts_address ON accounts(address);
CREATE INDEX idx_accounts_created_at ON accounts(created_at);

CREATE INDEX idx_validators_operator_address ON validators(operator_address);
CREATE INDEX idx_validators_status ON validators(status);
CREATE INDEX idx_validators_created_at ON validators(created_at);

CREATE INDEX idx_blocks_height ON blocks(height);
CREATE INDEX idx_blocks_hash ON blocks(hash);
CREATE INDEX idx_blocks_timestamp ON blocks(timestamp);
CREATE INDEX idx_blocks_proposer ON blocks(proposer_address);

CREATE INDEX idx_transactions_hash ON transactions(hash);
CREATE INDEX idx_transactions_block_height ON transactions(block_height);
CREATE INDEX idx_transactions_from_address ON transactions(from_address);
CREATE INDEX idx_transactions_to_address ON transactions(to_address);
CREATE INDEX idx_transactions_timestamp ON transactions(timestamp);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_status ON transactions(status);

CREATE INDEX idx_delegations_delegator ON delegations(delegator_address);
CREATE INDEX idx_delegations_validator ON delegations(validator_address);
CREATE INDEX idx_delegations_created_at ON delegations(created_at);

CREATE INDEX idx_validator_history_validator_id ON validator_history(validator_id);
CREATE INDEX idx_validator_history_block_height ON validator_history(block_height);
CREATE INDEX idx_validator_history_timestamp ON validator_history(timestamp);

CREATE INDEX idx_network_stats_block_height ON network_stats(block_height);
CREATE INDEX idx_network_stats_timestamp ON network_stats(timestamp);

CREATE INDEX idx_api_usage_endpoint ON api_usage(endpoint);
CREATE INDEX idx_api_usage_timestamp ON api_usage(timestamp);
CREATE INDEX idx_api_usage_ip_address ON api_usage(ip_address);

CREATE INDEX idx_user_sessions_session_id ON user_sessions(session_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_accounts_updated_at BEFORE UPDATE ON accounts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_validators_updated_at BEFORE UPDATE ON validators FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_delegations_updated_at BEFORE UPDATE ON delegations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create views for common queries
CREATE VIEW active_validators AS
SELECT 
    v.*,
    a.address as delegator_address_full
FROM validators v
LEFT JOIN accounts a ON a.address = v.delegator_address
WHERE v.status = 'active' AND v.jailed = FALSE;

CREATE VIEW validator_summary AS
SELECT 
    v.moniker,
    v.operator_address,
    v.commission_rate,
    v.total_delegation,
    v.voting_power,
    v.status,
    COUNT(d.id) as delegator_count,
    COALESCE(SUM(d.amount), 0) as total_delegated_amount
FROM validators v
LEFT JOIN delegations d ON d.validator_address = v.operator_address
GROUP BY v.id, v.moniker, v.operator_address, v.commission_rate, v.total_delegation, v.voting_power, v.status;

CREATE VIEW recent_blocks AS
SELECT 
    b.*,
    v.moniker as proposer_moniker
FROM blocks b
LEFT JOIN validators v ON v.consensus_address = b.proposer_address
ORDER BY b.height DESC
LIMIT 100;

CREATE VIEW network_overview AS
SELECT 
    (SELECT COUNT(*) FROM validators WHERE status = 'active') as active_validators,
    (SELECT COUNT(*) FROM validators) as total_validators,
    (SELECT COUNT(*) FROM accounts) as total_accounts,
    (SELECT MAX(height) FROM blocks) as latest_block_height,
    (SELECT COUNT(*) FROM transactions WHERE timestamp > NOW() - INTERVAL '24 hours') as transactions_24h,
    (SELECT SUM(total_delegation) FROM validators) as total_staked,
    (SELECT SUM(balance) FROM accounts) as total_balance;

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO tajeor;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO tajeor;
GRANT USAGE ON SCHEMA public TO tajeor;

-- Insert initial data (if needed)
-- This would be populated by the application from the blockchain data

COMMENT ON TABLE accounts IS 'Blockchain accounts with balances and metadata';
COMMENT ON TABLE validators IS 'Network validators with delegation information';
COMMENT ON TABLE blocks IS 'Blockchain blocks with transaction counts and metadata';
COMMENT ON TABLE transactions IS 'Individual transactions with details';
COMMENT ON TABLE delegations IS 'Delegation relationships between accounts and validators';
COMMENT ON TABLE validator_history IS 'Historical validator data for analytics';
COMMENT ON TABLE network_stats IS 'Network-wide statistics over time';
COMMENT ON TABLE api_usage IS 'API endpoint usage tracking';
COMMENT ON TABLE user_sessions IS 'User session management for authenticated access'; 