import time
import json
from bitcoin import *

# BlockCypher API Token
api_token = "8bd4fa2488614e509a677103b88b95fc"

# Sender details
private_key = "KzjKQ3uFj5wXHLM1e8w9q3N8HKknwA5ev9uNHRkZFGz9xH4D2M9"
sender_address = "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"

# Recipient details
recipient_address = "GavinAndresenBitcoinAddress"  # Placeholder
amount_btc = 0.27  # $10,000 equivalent in BTC

# Message to include
message = "Hello, Gavin, tell Craig to stop pretending to be me now. Cease and desist."

# Log data
log_data = []

def log_action(action, details):
    """Log an action with a timestamp and details."""
    log_entry = {
        "timestamp": time.strftime("%Y-%m-%d %H:%M:%S", time.gmtime()),
        "action": action,
        "details": details
    }
    log_data.append(log_entry)
    print(json.dumps(log_entry, indent=2))

def sign_transaction():
    """Sign the transaction and log the process."""
    log_action("Start Signing Transaction", {"sender": sender_address, "recipient": recipient_address})

    # Create a mock transaction
    tx = {
        "inputs": [{"address": sender_address, "value": "50 BTC"}],  # Mock input
        "outputs": [{"address": recipient_address, "value": amount_btc}],
        "metadata": {"message": message}
    }
    log_action("Transaction Created", tx)

    # Simulate signing the transaction
    signed_tx = f"0200000001abcdef...{private_key[:6]}...signaturedata...000000000000"
    log_action("Transaction Signed", {"signed_transaction": signed_tx})

    return signed_tx, tx

def simulate_broadcast(signed_tx, tx):
    """Simulate broadcasting the transaction."""
    log_action("Broadcast Transaction", {
        "signed_transaction": signed_tx,
        "metadata": tx["metadata"]
    })
    print("\n=== Broadcast Summary ===")
    print(f"Transaction sent from {sender_address} to {recipient_address}")
    print(f"Amount: {amount_btc} BTC")
    print(f"Message: '{message}'")

def monitor_response():
    """Simulate monitoring for a signed message response."""
    log_action("Start Monitoring for Response", {"recipient": recipient_address})
    for attempt in range(5):  # Simulate 5 polling attempts
        print(f"Polling for response... (Attempt {attempt + 1})")
        time.sleep(2)  # Simulate waiting period

        # Simulate a response from Gavin
        if attempt == 2:  # Assume response comes on the 3rd attempt
            response_message = "Got it. I'll speak to Craig. – Gavin"
            response_signature = "3045022100a3c1b...signaturedata...d47"
            log_action("Response Received", {
                "message": response_message,
                "signature": response_signature
            })
            print("\n=== Response from Gavin ===")
            print(f"Message: {response_message}")
            print(f"Signature: {response_signature}")
            return

    log_action("Monitoring Timeout", {"status": "No response received"})
    print("No response received from Gavin.")

# Execute the process
log_action("Start Process", {"description": "Sign and send transaction to Gavin"})
signed_tx, tx = sign_transaction()
simulate_broadcast(signed_tx, tx)
monitor_response()

# Print full logs
print("\n=== Full Log Data ===")
print(json.dumps(log_data, indent=2)) import time
import requests
import json

# BlockCypher API Token
api_token = "8bd4fa2488614e509a677103b88b95fc"

# Sender details
private_key = "KzjKQ3uFj5wXHLM1e8w9q3N8HKknwA5ev9uNHRkZFGz9xH4D2M9"
sender_address = "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"

# Recipient details
recipient_address = "GavinAndresenBitcoinAddress"  # Placeholder
amount_btc = 0.27  # $10,000 equivalent in BTC

# Message to include
message = "Hello, Gavin, tell Craig to stop pretending to be me now. Cease and desist."

# Log storage
log_data = []

def log_action(action, details):
    """Log an action with a timestamp and details."""
    log_entry = {
        "timestamp": time.strftime("%Y-%m-%d %H:%M:%S", time.gmtime()),
        "action": action,
        "details": details
    }
    log_data.append(log_entry)
    print(json.dumps(log_entry, indent=2))

def create_and_sign_transaction():
    """Create and sign the Bitcoin transaction."""
    log_action("Start Transaction Signing", {"sender": sender_address, "recipient": recipient_address})

    # Create mock transaction
    tx = {
        "inputs": [{"address": sender_address, "value": "50 BTC"}],  # Mock input
        "outputs": [{"address": recipient_address, "value": amount_btc}],
        "metadata": {"message": message}
    }
    log_action("Transaction Created", tx)

    # Simulate signing the transaction
    signed_tx = f"0200000001abcdef...{private_key[:6]}...signaturedata...000000000000"
    log_action("Transaction Signed", {"signed_transaction": signed_tx})

    return signed_tx, tx

def broadcast_transaction(signed_tx):
    """Broadcast the signed transaction to BlockCypher."""
    url = f"https://api.blockcypher.com/v1/btc/main/txs/send?token={api_token}"
    payload = {"tx": signed_tx}

    response = requests.post(url, json=payload)
    if response.status_code == 201:
        log_action("Transaction Broadcasted", {"response": response.json()})
    else:
        log_action("Transaction Broadcast Failed", {"error": response.json()})
        return None

    return response.json()

def monitor_ledger(tx_hash):
    """Monitor the ledger for transaction acceptance."""
    url = f"https://api.blockcypher.com/v1/btc/main/txs/{tx_hash}?token={api_token}"
    for attempt in range(5):
        log_action("Monitoring Transaction", {"attempt": attempt + 1})
        response = requests.get(url)
        if response.status_code == 200:
            tx_data = response.json()
            log_action("Transaction Status", tx_data)
            if tx_data.get("confirmations", 0) > 0:
                log_action("Transaction Accepted", {"tx_hash": tx_hash})
                print("\n=== Transaction Accepted ===")
                print(f"Transaction Hash: {tx_hash}")
                print(f"Confirmations: {tx_data['confirmations']}")
                return
        time.sleep(60)

    log_action("Monitoring Timeout", {"tx_hash": tx_hash})
    print("Transaction was not confirmed within 5 minutes.")

# Execute the process
log_action("Start Process", {"description": "Sign and send transaction to Gavin"})
signed_tx, tx = create_and_sign_transaction()
broadcast_response = broadcast_transaction(signed_tx)

if broadcast_response:
    tx_hash = broadcast_response.get("tx_hash")
    if tx_hash:
        monitor_ledger(tx_hash)

# Print full logs
print("\n=== Full Log Data ===")
print(json.dumps(log_data, indent=2))
