const express = require('express');
const cors = require('cors');
const fs = require('fs');
const path = require('path');
const { exec } = require('child_process');

const app = express();
const PORT = 3000;

// Middleware
app.use(cors());
app.use(express.json());
app.use(express.static('.'));

// Configuration
const TAJEOR_BINARY = '../build/tajeord.exe';
const NODE_HOME = process.env.USERPROFILE + '/.tajeor';

// Helper function to execute tajeord commands
function execTajeord(command) {
    return new Promise((resolve, reject) => {
        exec(`${TAJEOR_BINARY} ${command}`, { cwd: path.join(__dirname, '..') }, (error, stdout, stderr) => {
            if (error) {
                reject(error);
                return;
            }
            resolve(stdout);
        });
    });
}

// Helper function to read JSON files
function readJsonFile(filePath) {
    try {
        const data = fs.readFileSync(filePath, 'utf8');
        return JSON.parse(data);
    } catch (error) {
        return null;
    }
}

// Helper function to list files in directory
function listFiles(dirPath, extension = '.json') {
    try {
        const files = fs.readdirSync(dirPath);
        return files.filter(file => file.endsWith(extension));
    } catch (error) {
        return [];
    }
}

// API Routes

// Network statistics
app.get('/api/network-stats', async (req, res) => {
    try {
        // Get validator count
        const validatorsDir = path.join(NODE_HOME, 'validators');
        const validatorFiles = listFiles(validatorsDir);
        
        // Get genesis account count
        const genesisFile = path.join(NODE_HOME, 'config', 'genesis.json');
        const genesis = readJsonFile(genesisFile);
        
        let genesisAccounts = 0;
        if (genesis && genesis.app_state && genesis.app_state.auth && genesis.app_state.auth.accounts) {
            genesisAccounts = genesis.app_state.auth.accounts.length;
        }

        const stats = {
            validatorCount: validatorFiles.length,
            genesisAccounts: genesisAccounts,
            latestBlock: '12,345',
            totalSupply: '1,000,000,000 TJR',
            chainId: 'tajeor-1',
            networkStatus: 'online'
        };

        res.json(stats);
    } catch (error) {
        console.error('Error getting network stats:', error);
        res.status(500).json({ error: 'Failed to get network stats' });
    }
});

// Get validators
app.get('/api/validators', async (req, res) => {
    try {
        const validatorsDir = path.join(NODE_HOME, 'validators');
        const validatorFiles = listFiles(validatorsDir);
        
        const validators = [];
        
        for (const file of validatorFiles) {
            const validatorPath = path.join(validatorsDir, file);
            const validator = readJsonFile(validatorPath);
            
            if (validator) {
                // Get key information for delegator address
                const keysDir = path.join(NODE_HOME, 'keys');
                const keyPath = path.join(keysDir, validator.created_by + '.json');
                const keyInfo = readJsonFile(keyPath);
                
                validators.push({
                    moniker: validator.moniker,
                    operator_address: validator.operator_address,
                    commission_rate: validator.commission_rate,
                    min_self_delegation: validator.min_self_delegation,
                    self_delegation: validator.self_delegation,
                    total_delegation: validator.total_delegation,
                    voting_power: validator.voting_power,
                    status: validator.status,
                    created_by: validator.created_by,
                    delegator_address: keyInfo ? keyInfo.address : 'unknown'
                });
            }
        }
        
        res.json(validators);
    } catch (error) {
        console.error('Error getting validators:', error);
        res.status(500).json({ error: 'Failed to get validators' });
    }
});

// Get validator by moniker
app.get('/api/validator/:moniker', async (req, res) => {
    try {
        const { moniker } = req.params;
        const validatorPath = path.join(NODE_HOME, 'validators', moniker + '.json');
        const validator = readJsonFile(validatorPath);
        
        if (!validator) {
            return res.status(404).json({ error: 'Validator not found' });
        }
        
        // Get key information
        const keysDir = path.join(NODE_HOME, 'keys');
        const keyPath = path.join(keysDir, validator.created_by + '.json');
        const keyInfo = readJsonFile(keyPath);
        
        const result = {
            ...validator,
            delegator_address: keyInfo ? keyInfo.address : 'unknown'
        };
        
        res.json(result);
    } catch (error) {
        console.error('Error getting validator:', error);
        res.status(500).json({ error: 'Failed to get validator' });
    }
});

// Get account information
app.get('/api/account/:address', async (req, res) => {
    try {
        const { address } = req.params;
        
        // Check if address exists in genesis
        const genesisFile = path.join(NODE_HOME, 'config', 'genesis.json');
        const genesis = readJsonFile(genesisFile);
        
        let account = null;
        let balance = '0tjr';
        let accountNumber = null;
        let keyName = null;
        
        if (genesis && genesis.app_state) {
            // Check accounts
            if (genesis.app_state.auth && genesis.app_state.auth.accounts) {
                const foundAccount = genesis.app_state.auth.accounts.find(acc => acc.address === address);
                if (foundAccount) {
                    accountNumber = genesis.app_state.auth.accounts.indexOf(foundAccount);
                }
            }
            
            // Check balances
            if (genesis.app_state.bank && genesis.app_state.bank.balances) {
                const foundBalance = genesis.app_state.bank.balances.find(bal => bal.address === address);
                if (foundBalance) {
                    balance = foundBalance.coins;
                }
            }
        }
        
        // Find associated key name
        const keysDir = path.join(NODE_HOME, 'keys');
        const keyFiles = listFiles(keysDir);
        
        for (const file of keyFiles) {
            const keyPath = path.join(keysDir, file);
            const keyInfo = readJsonFile(keyPath);
            if (keyInfo && keyInfo.address === address) {
                keyName = keyInfo.name;
                break;
            }
        }
        
        if (accountNumber !== null || balance !== '0tjr') {
            account = {
                address: address,
                balance: balance,
                account_number: accountNumber,
                sequence: 0,
                key_name: keyName
            };
        }
        
        if (!account) {
            return res.status(404).json({ error: 'Account not found' });
        }
        
        res.json(account);
    } catch (error) {
        console.error('Error getting account:', error);
        res.status(500).json({ error: 'Failed to get account' });
    }
});

// Get all keys
app.get('/api/keys', async (req, res) => {
    try {
        const keysDir = path.join(NODE_HOME, 'keys');
        const keyFiles = listFiles(keysDir);
        
        const keys = [];
        
        for (const file of keyFiles) {
            const keyPath = path.join(keysDir, file);
            const keyInfo = readJsonFile(keyPath);
            if (keyInfo) {
                keys.push(keyInfo);
            }
        }
        
        res.json(keys);
    } catch (error) {
        console.error('Error getting keys:', error);
        res.status(500).json({ error: 'Failed to get keys' });
    }
});

// Get genesis information
app.get('/api/genesis', async (req, res) => {
    try {
        const genesisFile = path.join(NODE_HOME, 'config', 'genesis.json');
        const genesis = readJsonFile(genesisFile);
        
        if (!genesis) {
            return res.status(404).json({ error: 'Genesis file not found' });
        }
        
        res.json(genesis);
    } catch (error) {
        console.error('Error getting genesis:', error);
        res.status(500).json({ error: 'Failed to get genesis' });
    }
});

// Get block information (mock data for now)
app.get('/api/block/:height', async (req, res) => {
    try {
        const { height } = req.params;
        
        // Mock block data - in a real implementation, this would come from the blockchain
        const block = {
            height: height,
            hash: `0x${Math.random().toString(16).substr(2, 40)}`,
            timestamp: new Date().toISOString(),
            transactions: Math.floor(Math.random() * 50),
            proposer: 'cosmosvaloper1l2w3dalvtdd5hzlsumg4fvk58fu5l3xah3vy4p',
            size: Math.floor(Math.random() * 10000) + 1000
        };
        
        res.json(block);
    } catch (error) {
        console.error('Error getting block:', error);
        res.status(500).json({ error: 'Failed to get block' });
    }
});

// Get latest block
app.get('/api/block/latest', async (req, res) => {
    try {
        const block = {
            height: '12,345',
            hash: '0xdef789abc123456789def123456789abcdef123456',
            timestamp: new Date().toISOString(),
            transactions: 25,
            proposer: 'cosmosvaloper1puu5l8jxfst8euky84rmfjj2hy0uqznaeap0fc',
            size: 8547
        };
        
        res.json(block);
    } catch (error) {
        console.error('Error getting latest block:', error);
        res.status(500).json({ error: 'Failed to get latest block' });
    }
});

// Health check
app.get('/api/health', (req, res) => {
    res.json({ 
        status: 'healthy', 
        timestamp: new Date().toISOString(),
        version: '1.0.0',
        chain_id: 'tajeor-1'
    });
});

// Serve the explorer
app.get('/', (req, res) => {
    res.sendFile(path.join(__dirname, 'index.html'));
});

// Start server
app.listen(PORT, () => {
    console.log(`🌐 Tajeor Blockchain Explorer API Server running on http://localhost:${PORT}`);
    console.log(`📊 Explorer interface: http://localhost:${PORT}`);
    console.log(`🔗 API endpoints: http://localhost:${PORT}/api/`);
    console.log('');
    console.log('📋 Available API endpoints:');
    console.log('  GET /api/network-stats    - Network statistics');
    console.log('  GET /api/validators       - List all validators');
    console.log('  GET /api/validator/:name  - Get validator by moniker');
    console.log('  GET /api/account/:address - Get account information');
    console.log('  GET /api/keys            - List all keys');
    console.log('  GET /api/genesis         - Genesis file');
    console.log('  GET /api/block/:height   - Get block by height');
    console.log('  GET /api/block/latest    - Get latest block');
    console.log('  GET /api/health          - Health check');
});

module.exports = app; 