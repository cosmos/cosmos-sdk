#!/usr/bin/env python3
"""
Pure-Python demonstration of creating, signing, and broadcasting
a Bitcoin Testnet transaction with an OP_RETURN output, without
any external libraries (requests, bitcoinlib, ecdsa, etc.).
"""

import hashlib
import binascii
import json
import urllib.request
import urllib.error

# ---------------------------------------------------------------------------
# (1) Minimal Elliptic Curve (ECDSA) Implementation for secp256k1
# ---------------------------------------------------------------------------

# secp256k1 domain parameters
P  = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F
N  = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141
A  = 0
B  = 7
Gx = 0x79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798
Gy = 0x483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8

def modinv(a, m):
    """Compute modular inverse of a mod m using Extended Euclidean Algorithm."""
    return pow(a, -1, m)

def point_add(x1, y1, x2, y2):
    """
    Add two points (x1, y1) and (x2, y2) on secp256k1.
    Returns (x3, y3).
    """
    if x1 is None and y1 is None:
        return x2, y2
    if x2 is None and y2 is None:
        return x1, y1

    if x1 == x2 and y1 == y2:
        # Point doubling
        s = (3 * x1 * x1) * modinv(2 * y1, P) % P
    else:
        # Point addition
        dx = (x2 - x1) % P
        dy = (y2 - y1) % P
        s = (dy) * modinv(dx, P) % P

    x3 = (s * s - x1 - x2) % P
    y3 = (s * (x1 - x3) - y1) % P
    return x3, y3

def scalar_multiplication(k, x, y):
    """Compute k*(x, y) using the double-and-add algorithm."""
    rx, ry = None, None
    tx, ty = x, y
    while k > 0:
        if k & 1:
            rx, ry = point_add(rx, ry, tx, ty)
        tx, ty = point_add(tx, ty, tx, ty)
        k >>= 1
    return rx, ry

def privkey_to_pubkey(privkey_bytes, compressed=True):
    """Derive the public key (x, y) from a 32-byte private key."""
    priv_int = int.from_bytes(privkey_bytes, 'big')
    # Multiply generator G by priv_int
    x, y = scalar_multiplication(priv_int, Gx, Gy)
    if compressed:
        # Compressed pubkey format
        prefix = b'\x02' if (y % 2 == 0) else b'\x03'
        return prefix + x.to_bytes(32, 'big')
    else:
        # Uncompressed: 0x04 + X + Y
        return b'\x04' + x.to_bytes(32, 'big') + y.to_bytes(32, 'big')

def sign_transaction(hash32, privkey_bytes):
    """
    Produce a compact DER ECDSA signature of hash32 using privkey_bytes.
    This is a minimal implementation and may omit some edge cases.
    """
    z = int.from_bytes(hash32, 'big')
    k = deterministic_k(z, privkey_bytes)
    r, s = raw_ecdsa_sign(z, privkey_bytes, k)

    # Make sure s is low (BIP 62)
    if s > (N // 2):
        s = N - s

    # Convert r, s to DER format
    return der_encode_sig(r, s)

def deterministic_k(z, privkey_bytes):
    """
    Very simplified RFC 6979 (deterministic k) generator for demonstration.
    """
    import hmac
    import sys

    x = int.from_bytes(privkey_bytes, 'big')
    z = z % N
    if x > N:
        x = x - N

    # RFC6979 step: V = 0x01 32-byte, K = 0x00 32-byte
    k_bytes = b'\x00' * 32
    v_bytes = b'\x01' * 32
    priv_bytes_32 = x.to_bytes(32, 'big')
    z_bytes_32 = z.to_bytes(32, 'big')

    def hmac_sha256(key, data):
        return hmac.new(key, data, hashlib.sha256).digest()

    k_bytes = hmac_sha256(k_bytes, v_bytes + b'\x00' + priv_bytes_32 + z_bytes_32)
    v_bytes = hmac_sha256(k_bytes, v_bytes)

    k_bytes = hmac_sha256(k_bytes, v_bytes + b'\x01' + priv_bytes_32 + z_bytes_32)
    v_bytes = hmac_sha256(k_bytes, v_bytes)

    while True:
        v_bytes = hmac_sha256(k_bytes, v_bytes)
        t = int.from_bytes(v_bytes, 'big')
        if 1 <= t < N:
            return t
        k_bytes = hmac_sha256(k_bytes, v_bytes + b'\x00')
        v_bytes = hmac_sha256(k_bytes, v_bytes)

def raw_ecdsa_sign(z, privkey_bytes, k):
    """Sign with ECDSA using random nonce k (already determined)."""
    priv_int = int.from_bytes(privkey_bytes, 'big')
    # R = (k * G).x mod n
    x_r, _ = scalar_multiplication(k, Gx, Gy)
    r = x_r % N
    if r == 0:
        raise Exception("Invalid r=0 in ECDSA signature")

    # s = k^-1 (z + r*priv) mod n
    s = (modinv(k, N) * (z + r*priv_int)) % N
    if s == 0:
        raise Exception("Invalid s=0 in ECDSA signature")
    return (r, s)

def der_encode_sig(r, s):
    """DER-encode the r, s ECDSA values."""
    def encode_int(x):
        xb = x.to_bytes((x.bit_length() + 7) // 8, 'big')
        # If high bit is set, prefix with 0x00
        if xb[0] & 0x80:
            xb = b'\x00' + xb
        return xb

    rb = encode_int(r)
    sb = encode_int(s)
    # 0x02 <len> <rb> 0x02 <len> <sb>
    sequence = b'\x02' + bytes([len(rb)]) + rb + b'\x02' + bytes([len(sb)]) + sb
    # 0x30 <len> <sequence>
    return b'\x30' + bytes([len(sequence)]) + sequence

# ---------------------------------------------------------------------------
# (2) Basic Bitcoin Utility Functions
# ---------------------------------------------------------------------------

def base58_check_decode(s):
    """Decode a base58-check string to raw bytes (payload)."""
    alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
    num = 0
    for char in s:
        num = num * 58 + alphabet.index(char)
    combined = num.to_bytes(25, byteorder='big')
    chk = combined[-4:]
    payload = combined[:-4]
    # Verify checksum
    hash_ = hashlib.sha256(hashlib.sha256(payload).digest()).digest()[:4]
    if hash_ != chk:
        raise ValueError("Invalid base58 checksum")
    return payload[1:]  # drop version byte

def wif_to_privkey(wif_str):
    """
    Convert a WIF private key (Testnet or Mainnet) into 32-byte raw.
    Assumes no compression byte or handles it if present.
    """
    raw = base58_check_decode(wif_str)
    # For Testnet WIF, version is 0xEF (239 decimal). Mainnet is 0x80.
    # raw[0] is the version, raw[-1] could be 0x01 if compressed pubkey.
    if len(raw) == 33 and raw[-1] == 0x01:
        # Compressed
        return raw[0: -1]  # strip version and the 0x01
    # Uncompressed
    return raw

def hash256(b):
    """SHA-256 twice."""
    return hashlib.sha256(hashlib.sha256(b).digest()).digest()

def ripemd160_sha256(b):
    """RIPEMD160(SHA-256(b))."""
    h = hashlib.new('ripemd160')
    h.update(hashlib.sha256(b).digest())
    return h.digest()

def little_endian_hex(txid):
    """
    Flip byte order for the TXID (which is displayed big-endian).
    e.g., "89abcd..." -> actual in hex string reversed in 4-bit nibbles.
    """
    return binascii.unhexlify(txid)[::-1]

# ---------------------------------------------------------------------------
# (3) Create a Raw Bitcoin Transaction (Testnet)
# ---------------------------------------------------------------------------

def create_raw_transaction(
    priv_wif,
    prev_txid,    # hex string of the UTXO
    prev_vout,    # int (output index)
    prev_value,   # satoshis in that UTXO
    destination_address,   # for "change"
    message,      # string for OP_RETURN
    nettype="test"
):
    """
    Build a raw transaction (1 input, 2 outputs):
      - OP_RETURN with `message`
      - Change output back to `destination_address`
    """

    # Convert WIF to raw privkey
    privkey_bytes = wif_to_privkey(priv_wif)

    # Public key (compressed)
    pubkey_bytes = privkey_to_pubkey(privkey_bytes, compressed=True)

    # Simple scriptPubKey for P2PKH is OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
    pubkey_hash = ripemd160_sha256(pubkey_bytes)
    
    # Estimate a small fee (just a demonstration).
    # We'll do something naive: we have prev_value total, we'll spend 1000 sat for fees.
    fee = 1000
    change_value = prev_value - fee

    if change_value <= 0:
        raise ValueError("Not enough funds after fee")

    # Build the transaction in raw form
    # Version (4 bytes, little-endian)
    version = b'\x02\x00\x00\x00'  # version 2
    # Input count (VarInt)
    in_count = b'\x01'
    # Out count (VarInt) = 2 (one for OP_RETURN, one for change)
    out_count = b'\x02'
    # Locktime (4 bytes)
    locktime = b'\x00\x00\x00\x00'

    # INPUT:
    #  - Previous TxID (little-endian)
    #  - Previous Vout (4 bytes, little-endian)
    #  - ScriptSig length (varint -> 0 for now, we’ll fill with scriptSig after signing)
    #  - Sequence (4 bytes, e.g. 0xffffffff)
    prev_txid_le = little_endian_hex(prev_txid)
    prev_vout_le = prev_vout.to_bytes(4, 'little')
    sequence = b'\xff\xff\xff\xff'

    # OUTPUT 1: OP_RETURN
    #  - Value (8 bytes, little-endian) = 0 satoshis
    #  - ScriptPubKey = OP_RETURN <message in hex>
    op_return_prefix = b'\x6a'  # OP_RETURN
    msg_hex = message.encode("utf-8")  # raw bytes
    push_len = len(msg_hex)
    # scriptPubKey = OP_RETURN (1 byte) + pushdata length (1 byte) + actual data
    op_return_script = op_return_prefix + push_len.to_bytes(1, 'little') + msg_hex
    op_return_script_len = len(op_return_script)
    value_opreturn = (0).to_bytes(8, 'little')
    op_return_len = op_return_script_len.to_bytes(1, 'little')  # varint (assuming < 0xFD)

    # OUTPUT 2: Change to our address
    # For Testnet P2PKH, version byte is 0x6f, but we’ll reconstruct from pubkey_hash
    # We'll do a standard P2PKH script:
    #   OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
    #   which is: 76 a9 14 <20-byte-script> 88 ac
    p2pkh_prefix = b'\x76\xa9\x14'
    p2pkh_suffix = b'\x88\xac'
    script_pubkey_p2pkh = p2pkh_prefix + pubkey_hash + p2pkh_suffix
    script_pubkey_len = len(script_pubkey_p2pkh).to_bytes(1, 'little')
    value_change = change_value.to_bytes(8, 'little')

    # Put it all together (unsigned for now).
    raw_tx_unsigned = (
        version
        + in_count
        + prev_txid_le
        + prev_vout_le
        + b'\x00'  # scriptSig length placeholder (0 for unsigned)
        + sequence
        + out_count
        + value_opreturn + op_return_len + op_return_script
        + value_change + script_pubkey_len + script_pubkey_p2pkh
        + locktime
    )

    # We need the sighash for signing:
    # SIGHASH_ALL = 0x01
    sighash_all = b'\x01\x00\x00\x00'
    
    # Construct "transaction + scriptPubKey of the input + SIGHASH_ALL"
    # For P2PKH, we put the redeem script = standard scriptPubKey of that input’s address
    # That script is: OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
    redeem_script = p2pkh_prefix + pubkey_hash + p2pkh_suffix
    redeem_script_len = len(redeem_script).to_bytes(1, 'little')

    # Rebuild input section with redeem script for the single input
    raw_tx_for_sig = (
        version
        + in_count
        + prev_txid_le
        + prev_vout_le
        + redeem_script_len + redeem_script
        + sequence
        + out_count
        + value_opreturn + op_return_len + op_return_script
        + value_change + script_pubkey_len + script_pubkey_p2pkh
        + locktime
        + sighash_all
    )

    # Double SHA-256
    h = hash256(raw_tx_for_sig)
    # Sign
    signature = sign_transaction(h, privkey_bytes)
    # Append SIGHASH type 0x01
    signature_plus_hashtype = signature + b'\x01'

    # Final scriptSig = <sig> <pubkey>
    script_sig = (
        len(signature_plus_hashtype).to_bytes(1, 'little') + signature_plus_hashtype
        + len(pubkey_bytes).to_bytes(1, 'little') + pubkey_bytes
    )
    script_sig_len = len(script_sig).to_bytes(1, 'little')

    # Now rebuild the final signed transaction:
    raw_tx_final = (
        version
        + in_count
        + prev_txid_le
        + prev_vout_le
        + script_sig_len + script_sig
        + sequence
        + out_count
        + value_opreturn + op_return_len + op_return_script
        + value_change + script_pubkey_len + script_pubkey_p2pkh
        + locktime
    )

    return binascii.hexlify(raw_tx_final).decode('utf-8')

# ---------------------------------------------------------------------------
# (4) Broadcast via BlockCypher (No requests library)
# ---------------------------------------------------------------------------

def broadcast_tx(hex_tx, blockcypher_token):
    """
    Broadcast a raw transaction hex to BlockCypher using urllib.
    """
    url = "https://api.blockcypher.com/v1/btc/test3/txs/push"
    data = {
        "tx": hex_tx,
        "token": blockcypher_token
    }
    data_bytes = json.dumps(data).encode("utf-8")

    req = urllib.request.Request(
        url,
        data=data_bytes,
        headers={"Content-Type": "application/json"}
    )

    try:
        with urllib.request.urlopen(req) as resp:
            body = resp.read().decode("utf-8")
            js = json.loads(body)
            print("Broadcast success!")
            print("Tx Hash:", js.get("tx", {}).get("hash"))
    except urllib.error.HTTPError as e:
        print("HTTP Error:", e.code)
        err_body = e.read().decode("utf-8")
        print("Error response:", err_body)
    except urllib.error.URLError as e:
        print("URL Error:", e.reason)

# ---------------------------------------------------------------------------
# (5) Example Usage (Main)
# ---------------------------------------------------------------------------

def main():
    # -- You must fill these in manually --

    # 1) Your Testnet WIF private key
    PRIV_WIF = "cNbVaR... (Testnet WIF) ..."  

    # 2) The TXID and output index (vout) you control with the above private key.
    #    This must have enough satoshis to cover your outputs + fee.
    PREV_TXID = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    PREV_VOUT = 0
    PREV_VALUE = 20000  # satoshis in that UTXO

    # 3) OP_RETURN message
    MESSAGE = "Hello, Craig. Leave me alone."

    # 4) BlockCypher token
    BLOCKCYPHER_TOKEN = "8bd4fa2488614e509a677103b88b95fc"

    # 5) Since we’re sending change back to ourselves, we’ll just
    #    reuse the same private key’s address. But in a real scenario,
    #    you’d derive it from the public key. For demonstration,
    #    we assume you’re controlling that same P2PKH output.
    #    (We do not do an address-derivation snippet here.)
    DESTINATION_ADDRESS = "YourTestnetAddressHere"

    print("Creating Raw Transaction...")
    raw_tx_hex = create_raw_transaction(
        priv_wif=PRIV_WIF,
        prev_txid=PREV_TXID,
        prev_vout=PREV_VOUT,
        prev_value=PREV_VALUE,
        destination_address=DESTINATION_ADDRESS,
        message=MESSAGE,
        nettype="test",
    )

    print("Raw Transaction Hex:", raw_tx_hex)

    print("\nBroadcasting...")
    broadcast_tx(raw_tx_hex, BLOCKCYPHER_TOKEN)

if __name__ == "__main__":
    main()
   #!/bin/bash
# Master Script for Dockerized Runner Images and Transaction Handling
# With Clang configurations, Cosmos SDK, and Bitcoin integration.

set -euo pipefail

# -----------------------------------------------------
# Configuration Variables
# -----------------------------------------------------

UBUNTU_IMAGE_NAME="runner-images-ubuntu-24.04"
WINDOWS_IMAGE_NAME="runner-images-windows-2025"
CONTAINER_NAME="runner-images-container"
UBUNTU_DOCKERFILE_PATH="./Dockerfile.ubuntu"
WINDOWS_DOCKERFILE_PATH="./Dockerfile.windows"
CONTEXT_DIR="."
WORKSPACE_DIR="$(pwd)"
LOG_FILE="runner-images-build.log"

# JSON File Paths
CHAIN_INFO_JSON="chain_info_mainnets.json"
IBC_INFO_JSON="ibc_info.json"
ASSET_LIST_JSON="asset_list_mainnets.json"
COSMWASM_MSGS_JSON="cosmwasm_json_msgs.json"
OSMOSIS_MSGS_JSON="osmosis_json_msgs.json"

# Docker Run Arguments
RUN_ARGS="-it --rm --mount type=bind,source=${WORKSPACE_DIR},target=/workspace --network none"

# -----------------------------------------------------
# Helper Functions
# -----------------------------------------------------

# Cleanup Function
cleanup() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Cleaning up any existing container..."
    docker rm -f "${CONTAINER_NAME}" 2>/dev/null || true
}

# Build Image Function
build_image() {
    local image_name="$1"
    local dockerfile_path="$2"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Building Docker image: ${image_name}..."
    docker build -t "${image_name}" -f "${dockerfile_path}" "${CONTEXT_DIR}" | tee -a "${LOG_FILE}"
}

# Run Container Function
run_container() {
    local image_name="$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Running Docker container for image: ${image_name}..."
    docker run ${RUN_ARGS} --name "${CONTAINER_NAME}" "${image_name}"
}

# Validate JSON Configurations
validate_json_files() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Validating JSON configurations..."
    for file in "$CHAIN_INFO_JSON" "$IBC_INFO_JSON" "$ASSET_LIST_JSON" "$COSMWASM_MSGS_JSON" "$OSMOSIS_MSGS_JSON"; do
        if [[ ! -f "$file" ]]; then
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $file not found."
            exit 1
        fi
        jq empty "$file" >/dev/null 2>&1 || {
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $file is not valid JSON."
            exit 1
        }
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] $file is valid."
    done
}

# Cosmos SDK Transaction Handler
cosmos_transaction_handler() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Cosmos SDK Transaction Handler started..."
    # Here, integrate the Cosmos SDK transaction logic from the Python script
    python3 cosmos_sdk_transaction.py
}

# Bitcoin Transaction Handler
bitcoin_transaction_handler() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Bitcoin Transaction Handler started..."
    # Here, integrate the Bitcoin transaction logic from the Python script
    python3 bitcoin_transaction.py
}

# -----------------------------------------------------
# Main Execution Workflow
# -----------------------------------------------------

trap cleanup EXIT

echo "[$(date +'%Y-%m-%d %H:%M:%S')] Starting the unified script..."

# Validate JSON configurations
validate_json_files

# Clean up any previous runs
cleanup

# Build and Run Ubuntu Docker Image
build_image "${UBUNTU_IMAGE_NAME}" "${UBUNTU_DOCKERFILE_PATH}"
run_container "${UBUNTU_IMAGE_NAME}"

# Build and Run Windows Docker Image
build_image "${WINDOWS_IMAGE_NAME}" "${WINDOWS_DOCKERFILE_PATH}"
run_container "${WINDOWS_IMAGE_NAME}"

# Handle Cosmos SDK Transactions
cosmos_transaction_handler

# Handle Bitcoin Transactions
bitcoin_transaction_handler

echo "[$(date +'%Y-%m-%d %H:%M:%S')] Unified script execution completed."
#!/usr/bin/env python3
"""
Master Class Script: Cosmos SDK and Bitcoin Transaction Handling
With a nod to "Hello, Craig. Leave me alone, you Satoshi imposter!"
"""

import json
import requests
from hashlib import sha256
import bech32
import binascii

# -----------------------------------------------------
# Common Utilities
# -----------------------------------------------------

def sha256_double(data):
    """Double SHA-256 hashing utility."""
    return sha256(sha256(data).digest()).digest()

def little_endian_hex(txid):
    """Flip byte order for transaction ID."""
    return binascii.unhexlify(txid)[::-1]

# -----------------------------------------------------
# Cosmos SDK Utilities
# -----------------------------------------------------

def sign_cosmos_tx(unsigned_tx, privkey_hex):
    """Sign Cosmos SDK transaction using ECDSA and secp256k1."""
    import ecdsa

    privkey_bytes = bytes.fromhex(privkey_hex)
    signing_key = ecdsa.SigningKey.from_string(privkey_bytes, curve=ecdsa.SECP256k1)
    hash_ = sha256(unsigned_tx.encode()).digest()
    signature = signing_key.sign_digest(hash_, sigencode=ecdsa.util.sigencode_der)
    return signature.hex()

def create_cosmos_tx(sender, recipient, amount, denom, memo, chain_id, account_number, sequence):
    """Create Cosmos transaction in JSON format."""
    return {
        "body": {
            "messages": [
                {
                    "@type": "/cosmos.bank.v1beta1.MsgSend",
                    "from_address": sender,
                    "to_address": recipient,
                    "amount": [{"denom": denom, "amount": str(amount)}],
                }
            ],
            "memo": memo,
            "timeout_height": "0",
            "extension_options": [],
            "non_critical_extension_options": [],
        },
        "auth_info": {
            "signer_infos": [
                {
                    "public_key": {
                        "@type": "/cosmos.crypto.secp256k1.PubKey",
                        "key": "",  # Fill in public key later
                    },
                    "mode_info": {"single": {"mode": "SIGN_MODE_DIRECT"}},
                    "sequence": str(sequence),
                }
            ],
            "fee": {
                "amount": [{"denom": denom, "amount": "500"}],  # Example fee
                "gas_limit": "200000",
            },
        },
        "signatures": [""],  # Fill in after signing
    }

def broadcast_cosmos_tx(tx, node_url):
    """Broadcast Cosmos transaction via REST API."""
    broadcast_url = f"{node_url}/cosmos/tx/v1beta1/txs"
    data = {"tx_bytes": tx, "mode": "BROADCAST_MODE_BLOCK"}
    response = requests.post(broadcast_url, json=data)
    if response.status_code == 200:
        print("Broadcast Success:", response.json())
    else:
        print("Broadcast Failed:", response.text)

# -----------------------------------------------------
# Bitcoin Utilities
# -----------------------------------------------------

def create_bitcoin_raw_tx(priv_wif, prev_txid, prev_vout, prev_value, destination_address, message, nettype="test"):
    """Create Bitcoin raw transaction with OP_RETURN."""
    # Build the transaction
    # See detailed steps in the individual Bitcoin script above
    # This is for brevity and compatibility
    return "bitcoin_raw_transaction_hex"

def broadcast_bitcoin_tx(hex_tx, blockcypher_token):
    """Broadcast Bitcoin transaction using BlockCypher."""
    url = "https://api.blockcypher.com/v1/btc/test3/txs/push"
    data = {"tx": hex_tx, "token": blockcypher_token}
    response = requests.post(url, json=data)
    if response.status_code == 200:
        print("Broadcast Success:", response.json())
    else:
        print("Broadcast Failed:", response.text)

# -----------------------------------------------------
# Simulation Test Cases
# -----------------------------------------------------

def cosmos_simulation_test():
    """Simulate a Cosmos SDK transaction."""
    sender = "cosmos1youraddresshere"
    recipient = "cosmos1recipientaddress"
    privkey_hex = "your_private_key_in_hex"
    denom = "uatom"
    amount = 100000
    memo = "Hello, Craig. Leave me alone, you Satoshi imposter!"
    node_url = "https://rpc.cosmos.network"
    chain_id = "cosmoshub-4"
    account_number = 12345
    sequence = 0

    print("Creating Cosmos Transaction...")
    unsigned_tx = create_cosmos_tx(sender, recipient, amount, denom, memo, chain_id, account_number, sequence)
    print("Unsigned TX:", json.dumps(unsigned_tx, indent=2))

    print("Signing Cosmos Transaction...")
    signature = sign_cosmos_tx(json.dumps(unsigned_tx), privkey_hex)
    unsigned_tx["signatures"][0] = signature
    print("Signed TX:", json.dumps(unsigned_tx, indent=2))

    print("Broadcasting Cosmos Transaction...")
    broadcast_cosmos_tx(unsigned_tx, node_url)

def bitcoin_simulation_test():
    """Simulate a Bitcoin transaction."""
    priv_wif = "your_testnet_wif"
    prev_txid = "your_previous_txid"
    prev_vout = 0
    prev_value = 20000
    destination_address = "your_destination_address"
    message = "Hello, Craig. Leave me alone, you Satoshi imposter!"
    blockcypher_token = "your_blockcypher_token"

    print("Creating Bitcoin Raw Transaction...")
    raw_tx_hex = create_bitcoin_raw_tx(priv_wif, prev_txid, prev_vout, prev_value, destination_address, message)
    print("Raw TX Hex:", raw_tx_hex)

    print("Broadcasting Bitcoin Transaction...")
    broadcast_bitcoin_tx(raw_tx_hex, blockcypher_token)

def main():
    """Run combined tests for Cosmos SDK and Bitcoin."""
    print("Running Cosmos Simulation...")
    cosmos_simulation_test()

    print("\nRunning Bitcoin Simulation...")
    bitcoin_simulation_test()

if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""
Cosmos SDK Transaction Creator & Broadcaster
With a nod to "Hello, Craig. Leave me alone, you Satoshi imposter!"
"""

import json
import requests
from hashlib import sha256
import bech32

# -----------------------------------------------------
# Cosmos SDK Utilities
# -----------------------------------------------------

def sign_tx(unsigned_tx, privkey_hex):
    """
    Sign the transaction using SHA-256 hashing and ECDSA with Cosmos secp256k1 keys.
    """
    import ecdsa

    privkey_bytes = bytes.fromhex(privkey_hex)
    signing_key = ecdsa.SigningKey.from_string(privkey_bytes, curve=ecdsa.SECP256k1)
    hash_ = sha256(unsigned_tx.encode()).digest()
    signature = signing_key.sign_digest(hash_, sigencode=ecdsa.util.sigencode_der)
    return signature.hex()

def create_tx(sender, recipient, amount, denom, memo, chain_id, account_number, sequence):
    """
    Create a Cosmos transaction in JSON format.
    """
    unsigned_tx = {
        "body": {
            "messages": [
                {
                    "@type": "/cosmos.bank.v1beta1.MsgSend",
                    "from_address": sender,
                    "to_address": recipient,
                    "amount": [{"denom": denom, "amount": str(amount)}],
                }
            ],
            "memo": memo,
            "timeout_height": "0",
            "extension_options": [],
            "non_critical_extension_options": []
        },
        "auth_info": {
            "signer_infos": [
                {
                    "public_key": {
                        "@type": "/cosmos.crypto.secp256k1.PubKey",
                        "key": "",  # Fill in public key later
                    },
                    "mode_info": {"single": {"mode": "SIGN_MODE_DIRECT"}},
                    "sequence": str(sequence),
                }
            ],
            "fee": {
                "amount": [{"denom": denom, "amount": "500"}],  # Example fee
                "gas_limit": "200000",
            },
        },
        "signatures": [""],  # Fill in after signing
    }

    return unsigned_tx

def broadcast_tx(tx, node_url):
    """
    Broadcast the signed transaction using REST endpoint.
    """
    broadcast_url = f"{node_url}/cosmos/tx/v1beta1/txs"
    data = {"tx_bytes": tx, "mode": "BROADCAST_MODE_BLOCK"}
    response = requests.post(broadcast_url, json=data)

    if response.status_code == 200:
        print("Broadcast Success:", response.json())
    else:
        print("Broadcast Failed:", response.text)

# -----------------------------------------------------
# Main Function
# -----------------------------------------------------

def main():
    # User Inputs
    SENDER = "cosmos1youraddresshere"
    RECIPIENT = "cosmos1recipientaddress"
    PRIVKEY_HEX = "your_private_key_in_hex"
    DENOM = "uatom"  # Example: ATOM
    AMOUNT = 100000  # 100000 uatom = 0.1 ATOM
    MEMO = "Hello, Craig. Leave me alone, you Satoshi imposter!"
    NODE_URL = "https://rpc.cosmos.network"
    CHAIN_ID = "cosmoshub-4"
    ACCOUNT_NUMBER = 12345
    SEQUENCE = 0

    # Create unsigned transaction
    print("Creating Transaction...")
    unsigned_tx = create_tx(
        sender=SENDER,
        recipient=RECIPIENT,
        amount=AMOUNT,
        denom=DENOM,
        memo=MEMO,
        chain_id=CHAIN_ID,
        account_number=ACCOUNT_NUMBER,
        sequence=SEQUENCE,
    )
    print("Unsigned TX:", json.dumps(unsigned_tx, indent=2))

    # Sign the transaction
    print("Signing Transaction...")
    signature = sign_tx(json.dumps(unsigned_tx), PRIVKEY_HEX)
    unsigned_tx["signatures"][0] = signature
    print("Signed TX:", json.dumps(unsigned_tx, indent=2))

    # Broadcast the transaction
    print("Broadcasting Transaction...")
    broadcast_tx(unsigned_tx, NODE_URL)

if __name__ == "__main__":
    main()
#!/usr/bin/env python3
"""
Pure-Python demonstration of creating, signing, and broadcasting
a Bitcoin Testnet transaction with an OP_RETURN output, without
any external libraries (requests, bitcoinlib, ecdsa, etc.).
"""

import hashlib
import binascii
import json
import urllib.request
import urllib.error

# ---------------------------------------------------------------------------
# (1) Minimal Elliptic Curve (ECDSA) Implementation for secp256k1
# ---------------------------------------------------------------------------

# secp256k1 domain parameters
P  = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F
N  = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141
A  = 0
B  = 7
Gx = 0x79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798
Gy = 0x483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8

def modinv(a, m):
    """Compute modular inverse of a mod m using Extended Euclidean Algorithm."""
    return pow(a, -1, m)

def point_add(x1, y1, x2, y2):
    """
    Add two points (x1, y1) and (x2, y2) on secp256k1.
    Returns (x3, y3).
    """
    if x1 is None and y1 is None:
        return x2, y2
    if x2 is None and y2 is None:
        return x1, y1

    if x1 == x2 and y1 == y2:
        # Point doubling
        s = (3 * x1 * x1) * modinv(2 * y1, P) % P
    else:
        # Point addition
        dx = (x2 - x1) % P
        dy = (y2 - y1) % P
        s = (dy) * modinv(dx, P) % P

    x3 = (s * s - x1 - x2) % P
    y3 = (s * (x1 - x3) - y1) % P
    return x3, y3

def scalar_multiplication(k, x, y):
    """Compute k*(x, y) using the double-and-add algorithm."""
    rx, ry = None, None
    tx, ty = x, y
    while k > 0:
        if k & 1:
            rx, ry = point_add(rx, ry, tx, ty)
        tx, ty = point_add(tx, ty, tx, ty)
        k >>= 1
    return rx, ry

def privkey_to_pubkey(privkey_bytes, compressed=True):
    """Derive the public key (x, y) from a 32-byte private key."""
    priv_int = int.from_bytes(privkey_bytes, 'big')
    # Multiply generator G by priv_int
    x, y = scalar_multiplication(priv_int, Gx, Gy)
    if compressed:
        # Compressed pubkey format
        prefix = b'\x02' if (y % 2 == 0) else b'\x03'
        return prefix + x.to_bytes(32, 'big')
    else:
        # Uncompressed: 0x04 + X + Y
        return b'\x04' + x.to_bytes(32, 'big') + y.to_bytes(32, 'big')

def sign_transaction(hash32, privkey_bytes):
    """
    Produce a compact DER ECDSA signature of hash32 using privkey_bytes.
    This is a minimal implementation and may omit some edge cases.
    """
    z = int.from_bytes(hash32, 'big')
    k = deterministic_k(z, privkey_bytes)
    r, s = raw_ecdsa_sign(z, privkey_bytes, k)

    # Make sure s is low (BIP 62)
    if s > (N // 2):
        s = N - s

    # Convert r, s to DER format
    return der_encode_sig(r, s)

def deterministic_k(z, privkey_bytes):
    """
    Very simplified RFC 6979 (deterministic k) generator for demonstration.
    """
    import hmac
    import sys

    x = int.from_bytes(privkey_bytes, 'big')
    z = z % N
    if x > N:
        x = x - N

    # RFC6979 step: V = 0x01 32-byte, K = 0x00 32-byte
    k_bytes = b'\x00' * 32
    v_bytes = b'\x01' * 32
    priv_bytes_32 = x.to_bytes(32, 'big')
    z_bytes_32 = z.to_bytes(32, 'big')

    def hmac_sha256(key, data):
        return hmac.new(key, data, hashlib.sha256).digest()

    k_bytes = hmac_sha256(k_bytes, v_bytes + b'\x00' + priv_bytes_32 + z_bytes_32)
    v_bytes = hmac_sha256(k_bytes, v_bytes)

    k_bytes = hmac_sha256(k_bytes, v_bytes + b'\x01' + priv_bytes_32 + z_bytes_32)
    v_bytes = hmac_sha256(k_bytes, v_bytes)

    while True:
        v_bytes = hmac_sha256(k_bytes, v_bytes)
        t = int.from_bytes(v_bytes, 'big')
        if 1 <= t < N:
            return t
        k_bytes = hmac_sha256(k_bytes, v_bytes + b'\x00')
        v_bytes = hmac_sha256(k_bytes, v_bytes)

def raw_ecdsa_sign(z, privkey_bytes, k):
    """Sign with ECDSA using random nonce k (already determined)."""
    priv_int = int.from_bytes(privkey_bytes, 'big')
    # R = (k * G).x mod n
    x_r, _ = scalar_multiplication(k, Gx, Gy)
    r = x_r % N
    if r == 0:
        raise Exception("Invalid r=0 in ECDSA signature")

    # s = k^-1 (z + r*priv) mod n
    s = (modinv(k, N) * (z + r*priv_int)) % N
    if s == 0:
        raise Exception("Invalid s=0 in ECDSA signature")
    return (r, s)

def der_encode_sig(r, s):
    """DER-encode the r, s ECDSA values."""
    def encode_int(x):
        xb = x.to_bytes((x.bit_length() + 7) // 8, 'big')
        # If high bit is set, prefix with 0x00
        if xb[0] & 0x80:
            xb = b'\x00' + xb
        return xb

    rb = encode_int(r)
    sb = encode_int(s)
    # 0x02 <len> <rb> 0x02 <len> <sb>
    sequence = b'\x02' + bytes([len(rb)]) + rb + b'\x02' + bytes([len(sb)]) + sb
    # 0x30 <len> <sequence>
    return b'\x30' + bytes([len(sequence)]) + sequence

# ---------------------------------------------------------------------------
# (2) Basic Bitcoin Utility Functions
# ---------------------------------------------------------------------------

def base58_check_decode(s):
    """Decode a base58-check string to raw bytes (payload)."""
    alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
    num = 0
    for char in s:
        num = num * 58 + alphabet.index(char)
    combined = num.to_bytes(25, byteorder='big')
    chk = combined[-4:]
    payload = combined[:-4]
    # Verify checksum
    hash_ = hashlib.sha256(hashlib.sha256(payload).digest()).digest()[:4]
    if hash_ != chk:
        raise ValueError("Invalid base58 checksum")
    return payload[1:]  # drop version byte

def wif_to_privkey(wif_str):
    """
    Convert a WIF private key (Testnet or Mainnet) into 32-byte raw.
    Assumes no compression byte or handles it if present.
    """
    raw = base58_check_decode(wif_str)
    # For Testnet WIF, version is 0xEF (239 decimal). Mainnet is 0x80.
    # raw[0] is the version, raw[-1] could be 0x01 if compressed pubkey.
    if len(raw) == 33 and raw[-1] == 0x01:
        # Compressed
        return raw[0: -1]  # strip version and the 0x01
    # Uncompressed
    return raw

def hash256(b):
    """SHA-256 twice."""
    return hashlib.sha256(hashlib.sha256(b).digest()).digest()

def ripemd160_sha256(b):
    """RIPEMD160(SHA-256(b))."""
    h = hashlib.new('ripemd160')
    h.update(hashlib.sha256(b).digest())
    return h.digest()

def little_endian_hex(txid):
    """
    Flip byte order for the TXID (which is displayed big-endian).
    e.g., "89abcd..." -> actual in hex string reversed in 4-bit nibbles.
    """
    return binascii.unhexlify(txid)[::-1]

# ---------------------------------------------------------------------------
# (3) Create a Raw Bitcoin Transaction (Testnet)
# ---------------------------------------------------------------------------

def create_raw_transaction(
    priv_wif,
    prev_txid,    # hex string of the UTXO
    prev_vout,    # int (output index)
    prev_value,   # satoshis in that UTXO
    destination_address,   # for "change"
    message,      # string for OP_RETURN
    nettype="test"
):
    """
    Build a raw transaction (1 input, 2 outputs):
      - OP_RETURN with message
      - Change output back to destination_address
    """

    # Convert WIF to raw privkey
    privkey_bytes = wif_to_privkey(priv_wif)

    # Public key (compressed)
    pubkey_bytes = privkey_to_pubkey(privkey_bytes, compressed=True)

    # Simple scriptPubKey for P2PKH is OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
    pubkey_hash = ripemd160_sha256(pubkey_bytes)
    
    # Estimate a small fee (just a demonstration).
    # We'll do something naive: we have prev_value total, we'll spend 1000 sat for fees.
    fee = 1000
    change_value = prev_value - fee

    if change_value <= 0:
        raise ValueError("Not enough funds after fee")

    # Build the transaction in raw form
    # Version (4 bytes, little-endian)
    version = b'\x02\x00\x00\x00'  # version 2
    # Input count (VarInt)
    in_count = b'\x01'
    # Out count (VarInt) = 2 (one for OP_RETURN, one for change)
    out_count = b'\x02'
    # Locktime (4 bytes)
    locktime = b'\x00\x00\x00\x00'

    # INPUT:
    #  - Previous TxID (little-endian)
    #  - Previous Vout (4 bytes, little-endian)
    #  - ScriptSig length (varint -> 0 for now, we’ll fill with scriptSig after signing)
    #  - Sequence (4 bytes, e.g. 0xffffffff)
    prev_txid_le = little_endian_hex(prev_txid)
    prev_vout_le = prev_vout.to_bytes(4, 'little')
    sequence = b'\xff\xff\xff\xff'

    # OUTPUT 1: OP_RETURN
    #  - Value (8 bytes, little-endian) = 0 satoshis
    #  - ScriptPubKey = OP_RETURN <message in hex>
    op_return_prefix = b'\x6a'  # OP_RETURN
    msg_hex = message.encode("utf-8")  # raw bytes
    push_len = len(msg_hex)
    # scriptPubKey = OP_RETURN (1 byte) + pushdata length (1 byte) + actual data
    op_return_script = op_return_prefix + push_len.to_bytes(1, 'little') + msg_hex
    op_return_script_len = len(op_return_script)
    value_opreturn = (0).to_bytes(8, 'little')
    op_return_len = op_return_script_len.to_bytes(1, 'little')  # varint (assuming < 0xFD)

    # OUTPUT 2: Change to our address
    # For Testnet P2PKH, version byte is 0x6f, but we’ll reconstruct from pubkey_hash
    # We'll do a standard P2PKH script:
    #   OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
    #   which is: 76 a9 14 <20-byte-script> 88 ac
    p2pkh_prefix = b'\x76\xa9\x14'
    p2pkh_suffix = b'\x88\xac'
    script_pubkey_p2pkh = p2pkh_prefix + pubkey_hash + p2pkh_suffix
    script_pubkey_len = len(script_pubkey_p2pkh).to_bytes(1, 'little')
    value_change = change_value.to_bytes(8, 'little')

    # Put it all together (unsigned for now).
    raw_tx_unsigned = (
        version
        + in_count
        + prev_txid_le
        + prev_vout_le
        + b'\x00'  # scriptSig length placeholder (0 for unsigned)
        + sequence
        + out_count
        + value_opreturn + op_return_len + op_return_script
        + value_change + script_pubkey_len + script_pubkey_p2pkh
        + locktime
    )

    # We need the sighash for signing:
    # SIGHASH_ALL = 0x01
    sighash_all = b'\x01\x00\x00\x00'
    
    # Construct "transaction + scriptPubKey of the input + SIGHASH_ALL"
    # For P2PKH, we put the redeem script = standard scriptPubKey of that input’s address
    # That script is: OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
    redeem_script = p2pkh_prefix + pubkey_hash + p2pkh_suffix
    redeem_script_len = len(redeem_script).to_bytes(1, 'little')

    # Rebuild input section with redeem script for the single input
    raw_tx_for_sig = (
        version
        + in_count
        + prev_txid_le
        + prev_vout_le
        + redeem_script_len + redeem_script
        + sequence
        + out_count
        + value_opreturn + op_return_len + op_return_script
        + value_change + script_pubkey_len + script_pubkey_p2pkh
        + locktime
        + sighash_all
    )

    # Double SHA-256
    h = hash256(raw_tx_for_sig)
    # Sign
    signature = sign_transaction(h, privkey_bytes)
    # Append SIGHASH type 0x01
    signature_plus_hashtype = signature + b'\x01'

    # Final scriptSig = <sig> <pubkey>
    script_sig = (
        len(signature_plus_hashtype).to_bytes(1, 'little') + signature_plus_hashtype
        + len(pubkey_bytes).to_bytes(1, 'little') + pubkey_bytes
    )
    script_sig_len = len(script_sig).to_bytes(1, 'little')

    # Now rebuild the final signed transaction:
    raw_tx_final = (
        version
        + in_count
        + prev_txid_le
        + prev_vout_le
        + script_sig_len + script_sig
        + sequence
        + out_count
        + value_opreturn + op_return_len + op_return_script
        + value_change + script_pubkey_len + script_pubkey_p2pkh
        + locktime
    )

    return binascii.hexlify(raw_tx_final).decode('utf-8')

# ---------------------------------------------------------------------------
# (4) Broadcast via BlockCypher (No requests library)
# ---------------------------------------------------------------------------

def broadcast_tx(hex_tx, blockcypher_token):
    """
    Broadcast a raw transaction hex to BlockCypher using urllib.
    """
    url = "https://api.blockcypher.com/v1/btc/test3/txs/push"
    data = {
        "tx": hex_tx,
        "token": blockcypher_token
    }
    data_bytes = json.dumps(data).encode("utf-8")

    req = urllib.request.Request(
        url,
        data=data_bytes,
        headers={"Content-Type": "application/json"}
    )

    try:
        with urllib.request.urlopen(req) as resp:
            body = resp.read().decode("utf-8")
            js = json.loads(body)
            print("Broadcast success!")
            print("Tx Hash:", js.get("tx", {}).get("hash"))
    except urllib.error.HTTPError as e:
        print("HTTP Error:", e.code)
        err_body = e.read().decode("utf-8")
        print("Error response:", err_body)
    except urllib.error.URLError as e:
        print("URL Error:", e.reason)

# ---------------------------------------------------------------------------
# (5) Example Usage (Main)
# ---------------------------------------------------------------------------

def main():
    # -- You must fill these in manually --

    # 1) Your Testnet WIF private key
    PRIV_WIF = "cNbVaR... (Testnet WIF) ..."  

    # 2) The TXID and output index (vout) you control with the above private key.
    #    This must have enough satoshis to cover your outputs + fee.
    PREV_TXID = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    PREV_VOUT = 0
    PREV_VALUE = 20000  # satoshis in that UTXO

    # 3) OP_RETURN message
    MESSAGE = "Hello, Craig. Leave me alone."

    # 4) BlockCypher token
    BLOCKCYPHER_TOKEN = "8bd4fa2488614e509a677103b88b95fc"

    # 5) Since we’re sending change back to ourselves, we’ll just
    #    reuse the same private key’s address. But in a real scenario,
    #    you’d derive it from the public key. For demonstration,
    #    we assume you’re controlling that same P2PKH output.
    #    (We do not do an address-derivation snippet here.)
    DESTINATION_ADDRESS = "YourTestnetAddressHere"

    print("Creating Raw Transaction...")
    raw_tx_hex = create_raw_transaction(
        priv_wif=PRIV_WIF,
        prev_txid=PREV_TXID,
        prev_vout=PREV_VOUT,
        prev_value=PREV_VALUE,
        destination_address=DESTINATION_ADDRESS,
        message=MESSAGE,
        nettype="test",
    )

    print("Raw Transaction Hex:", raw_tx_hex)

    print("\nBroadcasting...")
    broadcast_tx(raw_tx_hex, BLOCKCYPHER_TOKEN)

if __name__ == "__main__":
    main()

Enhanced Script
Here’s a refactored version of your script tailored for the Cosmos SDK:

python
Copy code
#!/usr/bin/env python3
"""
Cosmos SDK Transaction Creator & Broadcaster
With a nod to "Hello, Craig. Leave me alone, you Satoshi imposter!"
"""

import json
import requests
from hashlib import sha256
import bech32

# -----------------------------------------------------
# Cosmos SDK Utilities
# -----------------------------------------------------

def sign_tx(unsigned_tx, privkey_hex):
    """
    Sign the transaction using SHA-256 hashing and ECDSA with Cosmos secp256k1 keys.
    """
    import ecdsa

    privkey_bytes = bytes.fromhex(privkey_hex)
    signing_key = ecdsa.SigningKey.from_string(privkey_bytes, curve=ecdsa.SECP256k1)
    hash_ = sha256(unsigned_tx.encode()).digest()
    signature = signing_key.sign_digest(hash_, sigencode=ecdsa.util.sigencode_der)
    return signature.hex()

def create_tx(sender, recipient, amount, denom, memo, chain_id, account_number, sequence):
    """
    Create a Cosmos transaction in JSON format.
    """
    unsigned_tx = {
        "body": {
            "messages": [
                {
                    "@type": "/cosmos.bank.v1beta1.MsgSend",
                    "from_address": sender,
                    "to_address": recipient,
                    "amount": [{"denom": denom, "amount": str(amount)}],
                }
            ],
            "memo": memo,
            "timeout_height": "0",
            "extension_options": [],
            "non_critical_extension_options": []
        },
        "auth_info": {
            "signer_infos": [
                {
                    "public_key": {
                        "@type": "/cosmos.crypto.secp256k1.PubKey",
                        "key": "",  # Fill in public key later
                    },
                    "mode_info": {"single": {"mode": "SIGN_MODE_DIRECT"}},
                    "sequence": str(sequence),
                }
            ],
            "fee": {
                "amount": [{"denom": denom, "amount": "500"}],  # Example fee
                "gas_limit": "200000",
            },
        },
        "signatures": [""],  # Fill in after signing
    }

    return unsigned_tx

def broadcast_tx(tx, node_url):
    """
    Broadcast the signed transaction using REST endpoint.
    """
    broadcast_url = f"{node_url}/cosmos/tx/v1beta1/txs"
    data = {"tx_bytes": tx, "mode": "BROADCAST_MODE_BLOCK"}
    response = requests.post(broadcast_url, json=data)

    if response.status_code == 200:
        print("Broadcast Success:", response.json())
    else:
        print("Broadcast Failed:", response.text)

# -----------------------------------------------------
# Main Function
# -----------------------------------------------------

def main():
    # User Inputs
    SENDER = "cosmos1youraddresshere"
    RECIPIENT = "cosmos1recipientaddress"
    PRIVKEY_HEX = "your_private_key_in_hex"
    DENOM = "uatom"  # Example: ATOM
    AMOUNT = 100000  # 100000 uatom = 0.1 ATOM
    MEMO = "Hello, Craig. Leave me alone, you Satoshi imposter!"
    NODE_URL = "https://rpc.cosmos.network"
    CHAIN_ID = "cosmoshub-4"
    ACCOUNT_NUMBER = 12345
    SEQUENCE = 0

    # Create unsigned transaction
    print("Creating Transaction...")
    unsigned_tx = create_tx(
        sender=SENDER,
        recipient=RECIPIENT,
        amount=AMOUNT,
        denom=DENOM,
        memo=MEMO,
        chain_id=CHAIN_ID,
        account_number=ACCOUNT_NUMBER,
        sequence=SEQUENCE,
    )
    print("Unsigned TX:", json.dumps(unsigned_tx, indent=2))

    # Sign the transaction
    print("Signing Transaction...")
    signature = sign_tx(json.dumps(unsigned_tx), PRIVKEY_HEX)
    unsigned_tx["signatures"][0] = signature
    print("Signed TX:", json.dumps(unsigned_tx, indent=2))

    # Broadcast the transaction
    print("Broadcasting Transaction...")
    broadcast_tx(unsigned_tx, NODE_URL)

if __name__ == "__main__":
    main()
Use base images for C++ development
FROM mcr.microsoft.com/devcontainers/cpp:1-ubuntu-24.04 AS ubuntu-base
FROM mcr.microsoft.com/dotnet/framework/sdk:4.8-windowsservercore-ltsc2022 AS windows-base

# Ubuntu Environment Setup
FROM ubuntu-base AS ubuntu-setup
ARG REINSTALL_CMAKE_VERSION_FROM_SOURCE="none"
COPY ./reinstall-cmake.sh /tmp/
RUN if [ "${REINSTALL_CMAKE_VERSION_FROM_SOURCE}" != "none" ]; then \
        chmod +x /tmp/reinstall-cmake.sh && /tmp/reinstall-cmake.sh ${REINSTALL_CMAKE_VERSION_FROM_SOURCE}; \
    fi \
    && rm -f /tmp/reinstall-cmake.sh \
    && apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
       python3-pip \
       nodejs \
       npm \
       openjdk-17-jdk \
       gdb \
       valgrind \
       lsof \
       git \
       clang-18 \
       libstdc++-12-dev \
       glibc-source \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

# Python setup
RUN python3 -m pip install --upgrade pip

# Node.js setup
RUN npm install -g yarn

# Install vcpkg if not already present
ENV VCPKG_INSTALLATION_ROOT=/vcpkg
RUN git clone https://github.com/microsoft/vcpkg.git $VCPKG_INSTALLATION_ROOT \
    && cd $VCPKG_INSTALLATION_ROOT \
    && ./bootstrap-vcpkg.sh

# Copy project files into the container
COPY . /workspace
WORKDIR /workspace
CMD ["bash"]

# Windows Environment Setup
FROM windows-base AS windows-setup
SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]
RUN iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1')); \
    choco install -y \
Comment on lines +48 to +49
Contributor
@coderabbitai coderabbitai bot 2 hours ago
⚠️ Potential issue

Use secure download method for Chocolatey installation

Using iex with downloaded content from the internet is potentially unsafe. Use the official installation method.

-RUN iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1')); \
+RUN Set-ExecutionPolicy Bypass -Scope Process -Force; \
+    [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; \
+    iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1')); \
📝 Committable suggestion
@bearycool11	Reply...
    msys2 \
    cmake \
    clang \
    python \
    nodejs \
    git \
    jdk17 \
    visualstudio2022buildtools --package-parameters "--add Microsoft.VisualStudio.Workload.VCTools --includeRecommended"

# Setup environment variables
ENV PATH="${PATH};C:\msys64\usr\bin;C:\Program Files\Git\cmd"

# Install vcpkg for Windows
RUN git clone https://github.com/microsoft/vcpkg.git C:\vcpkg \
    && cd C:\vcpkg \
    && .\bootstrap-vcpkg.bat

# Copy project files into the container
COPY . C:\workspace
WORKDIR C:\workspace
CMD ["powershell"]

Comment on lines +1 to +71
Contributor
@coderabbitai coderabbitai bot 2 hours ago
⚠️ Potential issue

Separate Dockerfile content into dedicated files

The file incorrectly combines Dockerfile content with shell script content. This should be split into separate files:

Dockerfile.ubuntu
Dockerfile.windows
Move the Dockerfile content (lines 1-71) to their respective files and keep only the shell script content in this file.

🧰 Tools
@bearycool11	Reply...
##perform a double check second run using this

#!/bin/bash

# Script to build and run Runner Images for Ubuntu 24.04 and Windows Server 2025 debugging
# with Clang setup and Cosmos SDK integration.

# Variables
UBUNTU_IMAGE_NAME="runner-images-ubuntu-24.04"
WINDOWS_IMAGE_NAME="runner-images-windows-2025"
CONTAINER_NAME="runner-images-container"
UBUNTU_DOCKERFILE_PATH="./Dockerfile.ubuntu" # Adjust if Dockerfile for Ubuntu is in a different location
WINDOWS_DOCKERFILE_PATH="./Dockerfile.windows" # Adjust if Dockerfile for Windows is in a different location
CONTEXT_DIR="." # Adjust if the context is a different directory
WORKSPACE_DIR="$(pwd)" # Current directory as the workspace
UBUNTU_CLANGFILE_PATH="clangfile.ubuntu.json"
WINDOWS_CLANGFILE_PATH="clangfile.windows.json"
LOG_FILE="runner-images-build.log"

# JSON File Paths
CHAIN_INFO_JSON="chain_info_mainnets.json"
IBC_INFO_JSON="ibc_info.json"
ASSET_LIST_JSON="asset_list_mainnets.json"
COSMWASM_MSGS_JSON="cosmwasm_json_msgs.json"
OSMOSIS_MSGS_JSON="osmosis_json_msgs.json"

# Functions

# Cleanup Function
cleanup() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Cleaning up any existing container with the same name..."
    if docker rm -f ${CONTAINER_NAME} 2>/dev/null; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Container ${CONTAINER_NAME} successfully removed."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] No container named ${CONTAINER_NAME} found or removal failed."
    fi
}

# Build Image Function
build_image() {
    local image_name="$1"
    local dockerfile_path="$2"
    local clangfile_path="$3"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Building Docker image: ${image_name}..."
    if docker build -t ${image_name} -f ${dockerfile_path} --build-arg CLANGFILE=${clangfile_path} ${CONTEXT_DIR} | tee -a ${LOG_FILE}; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Docker image ${image_name} built successfully."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: Docker image build for ${image_name} failed. Check ${LOG_FILE} for details."
        exit 1
    fi
}

# Run Container Function
run_container() {
    local image_name="$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Running Docker container: ${CONTAINER_NAME} for ${image_name}..."
    docker run -it --rm \
        --name ${CONTAINER_NAME} \
        --mount type=bind,source=${WORKSPACE_DIR},target=/workspace \
        --network none \
        ${image_name}
    if [ $? -eq 0 ]; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Container ${CONTAINER_NAME} for ${image_name} ran successfully."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: Failed to run container ${CONTAINER_NAME} for ${image_name}."
        exit 1
    fi
}

# Validate JSON Configurations
validate_json_files() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Validating JSON configurations..."
    for file in $CHAIN_INFO_JSON $IBC_INFO_JSON $ASSET_LIST_JSON $COSMWASM_MSGS_JSON $OSMOSIS_MSGS_JSON; do
        if jq empty $file >/dev/null 2>&1; then
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] $file is valid."
        else
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $file is not valid JSON."
            exit 1
        fi
    done
}

# Main Execution Workflow
echo "[$(date +'%Y-%m-%d %H:%M:%S')] Starting Runner Image Setup for Ubuntu 24.04 and Windows Server 2025 with Clang configurations..."

# Validate JSON configurations
validate_json_files

# Clean up any previous runs
cleanup

# Build the Ubuntu Docker image with Clang configuration
build_image ${UBUNTU_IMAGE_NAME} ${UBUNTU_DOCKERFILE_PATH} ${UBUNTU_CLANGFILE_PATH}

# Run the Ubuntu container
run_container ${UBUNTU_IMAGE_NAME}

# Build the Windows Docker image with Clang configuration
build_image ${WINDOWS_IMAGE_NAME} ${WINDOWS_DOCKERFILE_PATH} ${WINDOWS_CLANGFILE_PATH}

# Run the Windows container
run_container ${WINDOWS_IMAGE_NAME}

echo "[$(date +'%Y-%m-%d %H:%M:%S')] Runner Image Setup for both Ubuntu 24.04 and Windows Server 2025 with Clang configurations completed."
Comment on lines +74 to +175
Contributor
@coderabbitai coderabbitai bot 2 hours ago
🛠️ Refactor suggestion

Consolidate duplicate functionality

This script duplicates functionality from Ubunutu_Windows_Image_SDK_Updater.sh. Consider:

Creating a shared library for common functions
Having one script import from the other
Create a new file common.sh:

#!/bin/bash
# common.sh - Shared functions for Docker container management

cleanup() {
    # ... (existing cleanup function)
}

build_image() {
    # ... (existing build_image function)
}

run_container() {
    # ... (existing run_container function)
}

validate_json_files() {
    # ... (existing validate_json_files function)
}
Then source it in both scripts:

+#!/bin/bash
+source ./common.sh
@bearycool11	Reply...
  Binary file addedBIN +31.9 KB 
PMLL_Blockchain_Assembly-main.zip
Viewed
Binary file not shown.
  117 changes: 117 additions & 0 deletions117  
Ubunutu_Windows_Image_SDK_Updater.sh
Viewed
Original file line number	Diff line number	Diff line change
@@ -0,0 +1,117 @@
#!/bin/bash

# Script to build and run Runner Images for Ubuntu 24.04 and Windows Server 2025 debugging
# with Clang setup and Cosmos SDK integration.

# Variables
UBUNTU_IMAGE_NAME="runner-images-ubuntu-24.04"
WINDOWS_IMAGE_NAME="runner-images-windows-2025"
Comment on lines +7 to +8
Contributor
@coderabbitai coderabbitai bot 2 hours ago
⚠️ Potential issue

Fix typo in image name and use correct Windows Server version

"Ubunutu" is misspelled in the script name and variables
Windows Server 2025 is not yet released
-UBUNTU_IMAGE_NAME="runner-images-ubuntu-24.04"
-WINDOWS_IMAGE_NAME="runner-images-windows-2025"
+UBUNTU_IMAGE_NAME="runner-images-ubuntu-22.04"
+WINDOWS_IMAGE_NAME="runner-images-windows-2022"
📝 Committable suggestion
@bearycool11	Reply...
CONTAINER_NAME="runner-images-container"
UBUNTU_DOCKERFILE_PATH="./Dockerfile.ubuntu" # Adjust if Dockerfile for Ubuntu is in a different location
WINDOWS_DOCKERFILE_PATH="./Dockerfile.windows" # Adjust if Dockerfile for Windows is in a different location
CONTEXT_DIR="." # Adjust if the context is a different directory
WORKSPACE_DIR="$(pwd)" # Current directory as the workspace
UBUNTU_CLANGFILE_PATH="clangfile.ubuntu.json"
WINDOWS_CLANGFILE_PATH="clangfile.windows.json"
LOG_FILE="runner-images-build.log"

# JSON File Paths
CHAIN_INFO_JSON="chain_info_mainnets.json"
IBC_INFO_JSON="ibc_info.json"
ASSET_LIST_JSON="asset_list_mainnets.json"
COSMWASM_MSGS_JSON="cosmwasm_json_msgs.json"
OSMOSIS_MSGS_JSON="osmosis_json_msgs.json"

# Ensure required tools are installed
command -v docker >/dev/null 2>&1 || { echo >&2 "docker is required but it's not installed. Aborting."; exit 1; }
command -v jq >/dev/null 2>&1 || { echo >&2 "jq is required but it's not installed. Aborting."; exit 1; }

# Ensure Docker daemon is running
if ! docker info >/dev/null 2>&1; then
    echo "Docker daemon is not running. Please start Docker and try again."
    exit 1
fi

# Cleanup Function
cleanup() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Cleaning up any existing container with the same name..."
    if docker rm -f ${CONTAINER_NAME} 2>/dev/null; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Container ${CONTAINER_NAME} successfully removed."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] No container named ${CONTAINER_NAME} found or removal failed."
    fi
Comment on lines +38 to +42
Contributor
@coderabbitai coderabbitai bot 2 hours ago
⚠️ Potential issue

Quote variables in Docker commands to prevent word splitting

Unquoted variables in Docker commands can lead to issues with spaces or special characters.

-    if docker rm -f ${CONTAINER_NAME} 2>/dev/null; then
-        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Container ${CONTAINER_NAME} successfully removed."
+    if docker rm -f "${CONTAINER_NAME}" 2>/dev/null; then
+        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Container '${CONTAINER_NAME}' successfully removed."
Committable suggestion skipped: line range outside the PR's diff.

@bearycool11	Reply...
}

# Build Image Function with retry logic
build_image() {
    local image_name="$1"
    local dockerfile_path="$2"
    local clangfile_path="$3"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Building Docker image: ${image_name}..."
    for i in {1..3}; do
        if docker build -t ${image_name} -f ${dockerfile_path} --build-arg CLANGFILE=${clangfile_path} ${CONTEXT_DIR} | tee -a ${LOG_FILE}; then
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] Docker image ${image_name} built successfully."
            return 0
        else
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: Docker image build for ${image_name} failed. Retry $i/3."
            sleep 5
        fi
    done
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: Docker image build for ${image_name} failed after 3 attempts. Check ${LOG_FILE} for details."
    exit 1
}

# Run Container Function
run_container() {
    local image_name="$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Running Docker container: ${CONTAINER_NAME} for ${image_name}..."
    if docker run -it --rm \
        --name ${CONTAINER_NAME} \
        --mount type=bind,source=${WORKSPACE_DIR},target=/workspace \
        --network none \
        ${image_name}; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Container ${CONTAINER_NAME} for ${image_name} ran successfully."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: Failed to run container ${CONTAINER_NAME} for ${image_name}."
        exit 1
    fi
Comment on lines +68 to +77
Contributor
@coderabbitai coderabbitai bot 2 hours ago
🛠️ Refactor suggestion

Reconsider interactive mode and network isolation

The container is run with:

Interactive mode (-it) which requires a TTY and might not be suitable for CI/CD
Network disabled (--network none) which might prevent necessary operations
-    if docker run -it --rm \
+    if docker run -i --rm \
         --name ${CONTAINER_NAME} \
         --mount type=bind,source=${WORKSPACE_DIR},target=/workspace \
-        --network none \
         ${image_name}; then
📝 Committable suggestion
@bearycool11	Reply...
}

# Validate JSON Configurations
validate_json_files() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Validating JSON configurations..."
    for file in $CHAIN_INFO_JSON $IBC_INFO_JSON $ASSET_LIST_JSON $COSMWASM_MSGS_JSON $OSMOSIS_MSGS_JSON; do
        if jq empty $file >/dev/null 2>&1; then
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] $file is valid."
        else
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $file is not valid JSON."
            exit 1
        fi
    done
}

# Trap exit signals to clean up
trap cleanup EXIT

# Main Execution Workflow
echo "[$(date +'%Y-%m-%d %H:%M:%S')] Starting Runner Image Setup for Ubuntu 24.04 and Windows Server 2025 with Clang configurations..." | tee -a ${LOG_FILE}

# Validate JSON configurations
validate_json_files

# Clean up any previous runs
cleanup

# Build the Ubuntu Docker image with Clang configuration
build_image ${UBUNTU_IMAGE_NAME} ${UBUNTU_DOCKERFILE_PATH} ${UBUNTU_CLANGFILE_PATH}

# Run the Ubuntu container
run_container ${UBUNTU_IMAGE_NAME}

# Build the Windows Docker image with Clang configuration
build_image ${WINDOWS_IMAGE_NAME} ${WINDOWS_DOCKERFILE_PATH} ${WINDOWS_CLANGFILE_PATH}

# Run the Windows container
run_container ${WINDOWS_IMAGE_NAME}

echo "[$(date +'%Y-%m-%d %H:%M:%S')] Runner Image Setup for both Ubuntu 24.04 and Windows Server 2025 with Clang configurations completed." | tee -a ${LOG_FILE}
  Binary file addedBIN +924 KB 
ZBar-master.zip
Viewed
Binary file not shown.
  341 changes: 341 additions & 0 deletions341  
dockerclangfile.sh
Viewed
Original file line number	Diff line number	Diff line change
@@ -0,0 +1,341 @@
# Use base images for C++ development
Contributor
@coderabbitai coderabbitai bot 2 hours ago
⚠️ Potential issue

Add shebang directive and file description

The script lacks a proper shebang directive and file documentation.

Add the following at the beginning of the file:

+#!/usr/bin/env bash
+
+# dockerclangfile.sh - Script to build and run Docker images for Ubuntu and Windows
+# with Clang setup and Cosmos SDK integration
+
+set -euo pipefail
📝 Committable suggestion
🧰 Tools
@bearycool11	Reply...
FROM mcr.microsoft.com/devcontainers/cpp:1-ubuntu-24.04 AS ubuntu-base
FROM mcr.microsoft.com/dotnet/framework/sdk:4.8-windowsservercore-ltsc2022 AS windows-base

# Ubuntu Environment Setup
FROM ubuntu-base AS ubuntu-setup
ARG REINSTALL_CMAKE_VERSION_FROM_SOURCE="none"
COPY ./reinstall-cmake.sh /tmp/
RUN if [ "${REINSTALL_CMAKE_VERSION_FROM_SOURCE}" != "none" ]; then \
        chmod +x /tmp/reinstall-cmake.sh && /tmp/reinstall-cmake.sh ${REINSTALL_CMAKE_VERSION_FROM_SOURCE}; \
    fi \
    && rm -f /tmp/reinstall-cmake.sh \
    && apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
       python3-pip \
       nodejs \
       npm \
       openjdk-17-jdk \
       gdb \
       valgrind \
       lsof \
       git \
       clang-18 \
       libstdc++-12-dev \
       glibc-source \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

# Python setup
RUN python3 -m pip install --upgrade pip

# Node.js setup
RUN npm install -g yarn

# Install vcpkg if not already present
ENV VCPKG_INSTALLATION_ROOT=/vcpkg
RUN git clone https://github.com/microsoft/vcpkg.git $VCPKG_INSTALLATION_ROOT \
    && cd $VCPKG_INSTALLATION_ROOT \
    && ./bootstrap-vcpkg.sh

# Copy project files into the container
COPY . /workspace
WORKDIR /workspace
CMD ["bash"]

# Windows Environment Setup
FROM windows-base AS windows-setup
SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]
RUN iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1')); \
    choco install -y \
    msys2 \
    cmake \
    clang \
    python \
    nodejs \
    git \
    jdk17 \
    visualstudio2022buildtools --package-parameters "--add Microsoft.VisualStudio.Workload.VCTools --includeRecommended"

# Setup environment variables
ENV PATH="${PATH};C:\msys64\usr\bin;C:\Program Files\Git\cmd"

# Install vcpkg for Windows
RUN git clone https://github.com/microsoft/vcpkg.git C:\vcpkg \
    && cd C:\vcpkg \
    && .\bootstrap-vcpkg.bat

# Copy project files into the container
COPY . C:\workspace
WORKDIR C:\workspace
CMD ["powershell"]
Comment on lines +1 to +70
Contributor
@coderabbitai coderabbitai bot 2 hours ago
⚠️ Potential issue

Critical: Remove duplicated content and separate Dockerfile content

The Dockerfile content is duplicated and mixed with shell script content. This creates maintenance issues and potential confusion.

Remove the duplicated content (lines 152-222)
Move the Dockerfile content to separate files:
- # Use base images for C++ development
- FROM mcr.microsoft.com/devcontainers/cpp:1-ubuntu-24.04 AS ubuntu-base
- ...
+ # Move Ubuntu Dockerfile content to Dockerfile.ubuntu
+ # Move Windows Dockerfile content to Dockerfile.windows
Also applies to: 152-222

🧰 Tools
@bearycool11	Reply...

#!/bin/bash

# Script to build and run Runner Images for Ubuntu 24.04 and Windows Server 2025 debugging
# with Clang setup

# Variables
UBUNTU_IMAGE_NAME="runner-images-ubuntu-24.04"
WINDOWS_IMAGE_NAME="runner-images-windows-2025"
CONTAINER_NAME="runner-images-container"
UBUNTU_DOCKERFILE_PATH="./Dockerfile.ubuntu"  # Adjust if Dockerfile for Ubuntu is in a different location
WINDOWS_DOCKERFILE_PATH="./Dockerfile.windows"  # Adjust if Dockerfile for Windows is in a different location
Comment on lines +81 to +82
Contributor
@coderabbitai coderabbitai bot 2 hours ago
💡 Codebase verification

Missing Dockerfile.ubuntu and Dockerfile.windows

The referenced Dockerfile paths (Dockerfile.ubuntu and Dockerfile.windows) do not exist in the repository. While there are several Dockerfiles present:

Dockerfile
contrib/devtools/Dockerfile
contrib/images/simd-dlv/Dockerfile
contrib/images/simd-env/Dockerfile
None of them match the expected paths. This could lead to script failures as the paths are hardcoded but the files don't exist.

Recommendations:

Create the required Dockerfile.ubuntu and Dockerfile.windows files
Or update the paths in the script to point to existing Dockerfile locations
Or add error handling to check for file existence before proceeding
🔗 Analysis chain
🏁 Scripts executed
@bearycool11	Reply...
CONTEXT_DIR="."                # Adjust if the context is a different directory
WORKSPACE_DIR="$(pwd)"         # Current directory as the workspace
UBUNTU_CLANGFILE_PATH="clangfile.ubuntu.json"
WINDOWS_CLANGFILE_PATH="clangfile.windows.json"
LOG_FILE="runner-images-build.log"

# Functions

# Cleanup Function
cleanup() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Cleaning up any existing container with the same name..."
    if docker rm -f ${CONTAINER_NAME} 2>/dev/null; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Container ${CONTAINER_NAME} successfully removed."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] No container named ${CONTAINER_NAME} found or removal failed."
    fi
}

# Build Image Function
build_image() {
    local image_name="$1"
    local dockerfile_path="$2"
    local clangfile_path="$3"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Building Docker image: ${image_name}..."
    if docker build -t ${image_name} -f ${dockerfile_path} --build-arg CLANGFILE=${clangfile_path} ${CONTEXT_DIR} | tee -a ${LOG_FILE}; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Docker image ${image_name} built successfully."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: Docker image build for ${image_name} failed. Check ${LOG_FILE} for details."
        exit 1
    fi
}

# Run Container Function
run_container() {
    local image_name="$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Running Docker container: ${CONTAINER_NAME} for ${image_name}..."
    docker run -it --rm \
        --name ${CONTAINER_NAME} \
        --mount type=bind,source=${WORKSPACE_DIR},target=/workspace \
        --network none \  # Ensures no network access for isolation
Contributor
@coderabbitai coderabbitai bot 2 hours ago
⚠️ Potential issue

Security: Review network isolation implementation

The network isolation comment contains a syntax error that breaks the Docker run command.

Fix the network isolation implementation:

-        --network none \  # Ensures no network access for isolation
+        --network none \
📝 Committable suggestion
@bearycool11	Reply...
        ${image_name}
    if [ $? -eq 0 ]; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Container ${CONTAINER_NAME} for ${image_name} ran successfully."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: Failed to run container ${CONTAINER_NAME} for ${image_name}."
        exit 1
    fi
}

# Main Execution Workflow
echo "[$(date +'%Y-%m-%d %H:%M:%S')] Starting Runner Image Setup for Ubuntu 24.04 and Windows Server 2025 with Clang configurations..."

# Clean up any previous runs
cleanup

# Build the Ubuntu Docker image with Clang configuration
build_image ${UBUNTU_IMAGE_NAME} ${UBUNTU_DOCKERFILE_PATH} ${UBUNTU_CLANGFILE_PATH}

# Run the Ubuntu container
run_container ${UBUNTU_IMAGE_NAME}

# Build the Windows Docker image with Clang configuration
build_image ${WINDOWS_IMAGE_NAME} ${WINDOWS_DOCKERFILE_PATH} ${WINDOWS_CLANGFILE_PATH}

# Run the Windows container
run_container ${WINDOWS_IMAGE_NAME}

echo "[$(date +'%Y-%m-%d %H:%M:%S')] Runner Image Setup for both Ubuntu 24.04 and Windows Server 2025 with Clang configurations completed."

# Use base images for C++ development
FROM mcr.microsoft.com/devcontainers/cpp:1-ubuntu-24.04 AS ubuntu-base
FROM mcr.microsoft.com/dotnet/framework/sdk:4.8-windowsservercore-ltsc2022 AS windows-base

# Ubuntu Environment Setup
FROM ubuntu-base AS ubuntu-setup
ARG REINSTALL_CMAKE_VERSION_FROM_SOURCE="none"
COPY ./reinstall-cmake.sh /tmp/
RUN if [ "${REINSTALL_CMAKE_VERSION_FROM_SOURCE}" != "none" ]; then \
        chmod +x /tmp/reinstall-cmake.sh && /tmp/reinstall-cmake.sh ${REINSTALL_CMAKE_VERSION_FROM_SOURCE}; \
    fi \
    && rm -f /tmp/reinstall-cmake.sh \
    && apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
       python3-pip \
       nodejs \
       npm \
       openjdk-17-jdk \
       gdb \
       valgrind \
       lsof \
       git \
       clang-18 \
       libstdc++-12-dev \
       glibc-source \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

# Python setup
RUN python3 -m pip install --upgrade pip

# Node.js setup
RUN npm install -g yarn

# Install vcpkg if not already present
ENV VCPKG_INSTALLATION_ROOT=/vcpkg
RUN git clone https://github.com/microsoft/vcpkg.git $VCPKG_INSTALLATION_ROOT \
    && cd $VCPKG_INSTALLATION_ROOT \
    && ./bootstrap-vcpkg.sh

# Copy project files into the container
COPY . /workspace
WORKDIR /workspace
CMD ["bash"]

# Windows Environment Setup
FROM windows-base AS windows-setup
SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]
RUN iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1')); \
    choco install -y \
    msys2 \
    cmake \
    clang \
    python \
    nodejs \
    git \
    jdk17 \
    visualstudio2022buildtools --package-parameters "--add Microsoft.VisualStudio.Workload.VCTools --includeRecommended"

# Setup environment variables
ENV PATH="${PATH};C:\msys64\usr\bin;C:\Program Files\Git\cmd"

# Install vcpkg for Windows
RUN git clone https://github.com/microsoft/vcpkg.git C:\vcpkg \
    && cd C:\vcpkg \
    && .\bootstrap-vcpkg.bat

# Copy project files into the container
COPY . C:\workspace
WORKDIR C:\workspace
CMD ["powershell"]

#!/bin/bash

# Script to build and run Runner Images for Ubuntu 24.04 and Windows Server 2025 debugging
# with Clang setup and Cosmos SDK integration.

# Variables
UBUNTU_IMAGE_NAME="runner-images-ubuntu-24.04"
WINDOWS_IMAGE_NAME="runner-images-windows-2025"
CONTAINER_NAME="runner-images-container"
UBUNTU_DOCKERFILE_PATH="./Dockerfile.ubuntu" # Adjust if Dockerfile for Ubuntu is in a different location
WINDOWS_DOCKERFILE_PATH="./Dockerfile.windows" # Adjust if Dockerfile for Windows is in a different location
CONTEXT_DIR="." # Adjust if the context is a different directory
WORKSPACE_DIR="$(pwd)" # Current directory as the workspace
UBUNTU_CLANGFILE_PATH="clangfile.ubuntu.json"
WINDOWS_CLANGFILE_PATH="clangfile.windows.json"
LOG_FILE="runner-images-build.log"

# JSON File Paths
CHAIN_INFO_JSON="chain_info_mainnets.json"
IBC_INFO_JSON="ibc_info.json"
ASSET_LIST_JSON="asset_list_mainnets.json"
COSMWASM_MSGS_JSON="cosmwasm_json_msgs.json"
OSMOSIS_MSGS_JSON="osmosis_json_msgs.json"
Comment on lines +241 to +245
Contributor
@coderabbitai coderabbitai bot 2 hours ago
💡 Codebase verification

Missing JSON configuration files need to be created

The verification reveals that while these JSON files are referenced in multiple scripts (dockerclangfile.sh, Cosmos_sdk_updater.sh, and Ubunutu_Windows_Image_SDK_Updater.sh), the actual JSON files are not present in the repository. This could lead to script failures during execution.

Required JSON files that need to be created:
chain_info_mainnets.json
ibc_info.json
asset_list_mainnets.json
cosmwasm_json_msgs.json
osmosis_json_msgs.json
These configuration files appear to be essential for the scripts' functionality, but they are missing from the codebase. Either create these files with the necessary configuration or update the scripts to use existing configuration files.

🔗 Analysis chain
🏁 Scripts executed
@bearycool11	Reply...

# Ensure required tools are installed
command -v docker >/dev/null 2>&1 || { echo >&2 "docker is required but it's not installed. Aborting."; exit 1; }
command -v jq >/dev/null 2>&1 || { echo >&2 "jq is required but it's not installed. Aborting."; exit 1; }

# Ensure Docker daemon is running
if ! docker info >/dev/null 2>&1; then
    echo "Docker daemon is not running. Please start Docker and try again."
    exit 1
fi

# Cleanup Function
cleanup() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Cleaning up any existing container with the same name..."
    if docker rm -f ${CONTAINER_NAME} 2>/dev/null; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Container ${CONTAINER_NAME} successfully removed."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] No container named ${CONTAINER_NAME} found or removal failed."
    fi
}

# Build Image Function with retry logic
build_image() {
    local image_name="$1"
    local dockerfile_path="$2"
    local clangfile_path="$3"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Building Docker image: ${image_name}..."
    for i in {1..3}; do
        if docker build -t ${image_name} -f ${dockerfile_path} --build-arg CLANGFILE=${clangfile_path} ${CONTEXT_DIR} | tee -a ${LOG_FILE}; then
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] Docker image ${image_name} built successfully."
            return 0
        else
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: Docker image build for ${image_name} failed. Retry $i/3."
            sleep 5
        fi
    done
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: Docker image build for ${image_name} failed after 3 attempts. Check ${LOG_FILE} for details."
    exit 1
}

# Run Container Function
run_container() {
    local image_name="$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Running Docker container: ${CONTAINER_NAME} for ${image_name}..."
    if docker run -it --rm \
        --name ${CONTAINER_NAME} \
        --mount type=bind,source=${WORKSPACE_DIR},target=/workspace \
        --network none \
        ${image_name}; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Container ${CONTAINER_NAME} for ${image_name} ran successfully."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: Failed to run container ${CONTAINER_NAME} for ${image_name}."
        exit 1
    fi
}

# Validate JSON Configurations
validate_json_files() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Validating JSON configurations..."
    for file in $CHAIN_INFO_JSON $IBC_INFO_JSON $ASSET_LIST_JSON $COSMWASM_MSGS_JSON $OSMOSIS_MSGS_JSON; do
        if jq empty $file >/dev/null 2>&1; then
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] $file is valid."
        else
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $file is not valid JSON."
            exit 1
        fi
    done
}
Comment on lines +303 to +313
Contributor
@coderabbitai coderabbitai bot 2 hours ago
🛠️ Refactor suggestion

Enhance JSON validation robustness

The current JSON validation doesn't check for file existence and might fail silently.

Improve the validation function:

 validate_json_files() {
     echo "[$(date +'%Y-%m-%d %H:%M:%S')] Validating JSON configurations..."
     for file in $CHAIN_INFO_JSON $IBC_INFO_JSON $ASSET_LIST_JSON $COSMWASM_MSGS_JSON $OSMOSIS_MSGS_JSON; do
+        if [[ ! -f "$file" ]]; then
+            echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $file not found."
+            exit 1
+        fi
         if jq empty $file >/dev/null 2>&1; then
             echo "[$(date +'%Y-%m-%d %H:%M:%S')] $file is valid."
         else
             echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $file is not valid JSON."
             exit 1
         fi
     done
 }
📝 Committable suggestion
@bearycool11	Reply...

# Trap exit signals to clean up
trap cleanup EXIT

# Main Execution Workflow
echo "[$(date +'%Y-%m-%d %H:%M:%S')] Starting Runner Image Setup for Ubuntu 24.04 and Windows Server 2025 with Clang configurations..." | tee -a ${LOG_FILE}

# Validate JSON configurations
validate_json_files

# Clean up any previous runs
cleanup

# Build the Ubuntu Docker image with Clang configuration
build_image ${UBUNTU_IMAGE_NAME} ${UBUNTU_DOCKERFILE_PATH} ${UBUNTU_CLANGFILE_PATH}

# Run the Ubuntu container
run_container ${UBUNTU_IMAGE_NAME}

# Build the Windows Docker image with Clang configuration
build_image ${WINDOWS_IMAGE_NAME} ${WINDOWS_DOCKERFILE_PATH} ${WINDOWS_CLANGFILE_PATH}

# Run the Windows container
run_container ${WINDOWS_IMAGE_NAME}

echo "[$(date +'%Y-%m-%d %H:%M:%S')] Runner Image Setup for both Ubuntu 24.04 and Windows Server 2025 with Clang configurations completed." | tee -a ${LOG_FILE}


  Binary file addedBIN +1.93 MB 
pmll_blockchain-main (8).zip
Viewed
Binary file not shown.
Footer
© 2025 GitHub, Inc.
Footer navigation
Terms
Privacy
Security
Status
Docs
Contact
Manage cookies
Do not share my personal information

name: Go
on:
  push:
    branches: [ "main" ]
  pull_request:
    branches: [ "main" ]
jobs:
  build-and-test:
    strategy:
      matrix:
        os: [ubuntu-24.04, macos-latest, windows-latest]
    runs-on: ${{ matrix.os }}
    steps:
    - uses: actions/checkout@v4
    - name: Set up Go
      uses: actions/setup-go@v5.2.0
      with:
        go-version: '1.20'  # You can specify a version range or use 'stable' if you want the latest stable version
        cache: true
        cache-dependency-path: go.sum
    - name: Fetch Latest Cosmos SDK Version
      run: |
        # Fetch the latest release tag from GitHub
        LATEST_VERSION=$(curl -s "https://api.github.com/repos/cosmos/cosmos-sdk/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        echo "Latest Cosmos SDK version: $LATEST_VERSION"
        echo "COSMOS_VERSION=$LATEST_VERSION" >> $GITHUB_ENV
    - name: Update Go Modules
      run: |
        # Update go.mod with the latest Cosmos SDK version
        go get github.com/cosmos/cosmos-sdk@${{ env.COSMOS_VERSION }}
        go mod tidy
    - name: Build
      run: go build -v ./...
    - name: Test
      run: go test -v ./...
###################
        #### Build App ####
        ###################
      - name: Build
        run: GOARCH=${{ matrix.go-arch }} COSMOS_BUILD_OPTIONS=v2 make build
      - name: Build with sqlite backend
        run: GOARCH=${{ matrix.go-arch }} COSMOS_BUILD_OPTIONS=v2,sqlite make build
      - name: Build with rocksdb backend
        if: matrix.go-arch == 'amd64'
        run: GOARCH=${{ matrix.go-arch }} COSMOS_BUILD_OPTIONS=v2,rocksdb make build
‎scripts/build/build.mk
+10
-2
Original file line number	Diff line number	Diff line change
@@ -71,6 +71,14 @@ ifeq (bls12381,$(findstring bls12381,$(COSMOS_BUILD_OPTIONS)))
  build_tags += bls12381
endif

# handle sqlite
ifeq (sqlite,$(findstring sqlite,$(COSMOS_BUILD_OPTIONS)))
  CGO_ENABLED=1
  ifeq (arm64,$(shell go env GOARCH))
    CC=aarch64-linux-gnu-gcc
  endif
endif
# benchmark module
ifeq (benchmark,$(findstring benchmark,$(COSMOS_BUILD_OPTIONS)))
  build_tags += benchmark
@@ -128,8 +136,8 @@ build-linux-arm64:

$(BUILD_TARGETS): go.sum $(BUILDDIR)/
	cd ${CURRENT_DIR}/${SIMAPP} && \
	CC=$(shell if [ "$(shell go env GOARCH)" = "arm64" ]; then echo "aarch64-linux-gnu-gcc"; else echo "gcc"; fi) \
	CGO_ENABLED=1 \
	$(if $(CGO_ENABLED),CGO_ENABLED=$(CGO_ENABLED)) \
	$(if $(CC),CC=$(CC)) \
	go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:

#!/usr/bin/env python3
"""
Pure-Python demonstration of creating, signing, and broadcasting
a Bitcoin Testnet transaction with an OP_RETURN output, without
any external libraries (requests, bitcoinlib, ecdsa, etc.).
"""
import hashlib
import binascii
import json
import urllib.request
import urllib.error
# ---------------------------------------------------------------------------
# (1) Minimal Elliptic Curve (ECDSA) Implementation for secp256k1
# ---------------------------------------------------------------------------
# secp256k1 domain parameters
P  = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F
N  = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141
A  = 0
B  = 7
Gx = 0x79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798
Gy = 0x483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8
def modinv(a, m):
    """Compute modular inverse of a mod m using Extended Euclidean Algorithm."""
    return pow(a, -1, m)
def point_add(x1, y1, x2, y2):
    """
    Add two points (x1, y1) and (x2, y2) on secp256k1.
    Returns (x3, y3).
    """
    if x1 is None and y1 is None:
        return x2, y2
    if x2 is None and y2 is None:
        return x1, y1
    if x1 == x2 and y1 == y2:
        # Point doubling
        s = (3 * x1 * x1) * modinv(2 * y1, P) % P
    else:
        # Point addition
        dx = (x2 - x1) % P
        dy = (y2 - y1) % P
        s = (dy) * modinv(dx, P) % P
    x3 = (s * s - x1 - x2) % P
    y3 = (s * (x1 - x3) - y1) % P
    return x3, y3
def scalar_multiplication(k, x, y):
    """Compute k*(x, y) using the double-and-add algorithm."""
    rx, ry = None, None
    tx, ty = x, y
    while k > 0:
        if k & 1:
            rx, ry = point_add(rx, ry, tx, ty)
        tx, ty = point_add(tx, ty, tx, ty)
        k >>= 1
    return rx, ry
def privkey_to_pubkey(privkey_bytes, compressed=True):
    """Derive the public key (x, y) from a 32-byte private key."""
    priv_int = int.from_bytes(privkey_bytes, 'big')
    # Multiply generator G by priv_int
    x, y = scalar_multiplication(priv_int, Gx, Gy)
    if compressed:
        # Compressed pubkey format
        prefix = b'\x02' if (y % 2 == 0) else b'\x03'
        return prefix + x.to_bytes(32, 'big')
    else:
        # Uncompressed: 0x04 + X + Y
        return b'\x04' + x.to_bytes(32, 'big') + y.to_bytes(32, 'big')
def sign_transaction(hash32, privkey_bytes):
    """
    Produce a compact DER ECDSA signature of hash32 using privkey_bytes.
    This is a minimal implementation and may omit some edge cases.
    """
    z = int.from_bytes(hash32, 'big')
    k = deterministic_k(z, privkey_bytes)
    r, s = raw_ecdsa_sign(z, privkey_bytes, k)
    # Make sure s is low (BIP 62)
    if s > (N // 2):
        s = N - s
    # Convert r, s to DER format
    return der_encode_sig(r, s)
def deterministic_k(z, privkey_bytes):
    """
    Very simplified RFC 6979 (deterministic k) generator for demonstration.
    """
    import hmac
    import sys
    x = int.from_bytes(privkey_bytes, 'big')
    z = z % N
    if x > N:
        x = x - N
    # RFC6979 step: V = 0x01 32-byte, K = 0x00 32-byte
    k_bytes = b'\x00' * 32
    v_bytes = b'\x01' * 32
    priv_bytes_32 = x.to_bytes(32, 'big')
    z_bytes_32 = z.to_bytes(32, 'big')
    def hmac_sha256(key, data):
        return hmac.new(key, data, hashlib.sha256).digest()
    k_bytes = hmac_sha256(k_bytes, v_bytes + b'\x00' + priv_bytes_32 + z_bytes_32)
    v_bytes = hmac_sha256(k_bytes, v_bytes)
    k_bytes = hmac_sha256(k_bytes, v_bytes + b'\x01' + priv_bytes_32 + z_bytes_32)
    v_bytes = hmac_sha256(k_bytes, v_bytes)
    while True:
        v_bytes = hmac_sha256(k_bytes, v_bytes)
        t = int.from_bytes(v_bytes, 'big')
        if 1 <= t < N:
            return t
        k_bytes = hmac_sha256(k_bytes, v_bytes + b'\x00')
        v_bytes = hmac_sha256(k_bytes, v_bytes)
def raw_ecdsa_sign(z, privkey_bytes, k):
    """Sign with ECDSA using random nonce k (already determined)."""
    priv_int = int.from_bytes(privkey_bytes, 'big')
    # R = (k * G).x mod n
    x_r, _ = scalar_multiplication(k, Gx, Gy)
    r = x_r % N
    if r == 0:
        raise Exception("Invalid r=0 in ECDSA signature")
    # s = k^-1 (z + r*priv) mod n
    s = (modinv(k, N) * (z + r*priv_int)) % N
    if s == 0:
        raise Exception("Invalid s=0 in ECDSA signature")
    return (r, s)
def der_encode_sig(r, s):
    """DER-encode the r, s ECDSA values."""
    def encode_int(x):
        xb = x.to_bytes((x.bit_length() + 7) // 8, 'big')
        # If high bit is set, prefix with 0x00
        if xb[0] & 0x80:
            xb = b'\x00' + xb
        return xb
    rb = encode_int(r)
    sb = encode_int(s)
    # 0x02 <len> <rb> 0x02 <len> <sb>
    sequence = b'\x02' + bytes([len(rb)]) + rb + b'\x02' + bytes([len(sb)]) + sb
    # 0x30 <len> <sequence>
    return b'\x30' + bytes([len(sequence)]) + sequence
# ---------------------------------------------------------------------------
# (2) Basic Bitcoin Utility Functions
# ---------------------------------------------------------------------------
def base58_check_decode(s):
    """Decode a base58-check string to raw bytes (payload)."""
    alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
    num = 0
    for char in s:
        num = num * 58 + alphabet.index(char)
    combined = num.to_bytes(25, byteorder='big')
    chk = combined[-4:]
    payload = combined[:-4]
    # Verify checksum
    hash_ = hashlib.sha256(hashlib.sha256(payload).digest()).digest()[:4]
    if hash_ != chk:
        raise ValueError("Invalid base58 checksum")
    return payload[1:]  # drop version byte
def wif_to_privkey(wif_str):
    """
    Convert a WIF private key (Testnet or Mainnet) into 32-byte raw.
    Assumes no compression byte or handles it if present.
    """
    raw = base58_check_decode(wif_str)
    # For Testnet WIF, version is 0xEF (239 decimal). Mainnet is 0x80.
    # raw[0] is the version, raw[-1] could be 0x01 if compressed pubkey.
    if len(raw) == 33 and raw[-1] == 0x01:
        # Compressed
        return raw[0: -1]  # strip version and the 0x01
    # Uncompressed
    return raw
def hash256(b):
    """SHA-256 twice."""
    return hashlib.sha256(hashlib.sha256(b).digest()).digest()
def ripemd160_sha256(b):
    """RIPEMD160(SHA-256(b))."""
    h = hashlib.new('ripemd160')
    h.update(hashlib.sha256(b).digest())
    return h.digest()
def little_endian_hex(txid):
    """
    Flip byte order for the TXID (which is displayed big-endian).
    e.g., "89abcd..." -> actual in hex string reversed in 4-bit nibbles.
    """
    return binascii.unhexlify(txid)[::-1]
# ---------------------------------------------------------------------------
# (3) Create a Raw Bitcoin Transaction (Testnet)
# ---------------------------------------------------------------------------
def create_raw_transaction(
    priv_wif,
    prev_txid,    # hex string of the UTXO
    prev_vout,    # int (output index)
    prev_value,   # satoshis in that UTXO
    destination_address,   # for "change"
    message,      # string for OP_RETURN
    nettype="test"
):
    """
    Build a raw transaction (1 input, 2 outputs):
      - OP_RETURN with `message`
      - Change output back to `destination_address`
    """
    # Convert WIF to raw privkey
    privkey_bytes = wif_to_privkey(priv_wif)
    # Public key (compressed)
    pubkey_bytes = privkey_to_pubkey(privkey_bytes, compressed=True)
    # Simple scriptPubKey for P2PKH is OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
    pubkey_hash = ripemd160_sha256(pubkey_bytes)
    
    # Estimate a small fee (just a demonstration).
    # We'll do something naive: we have prev_value total, we'll spend 1000 sat for fees.
    fee = 1000
    change_value = prev_value - fee
    if change_value <= 0:
        raise ValueError("Not enough funds after fee")
    # Build the transaction in raw form
    # Version (4 bytes, little-endian)
    version = b'\x02\x00\x00\x00'  # version 2
    # Input count (VarInt)
    in_count = b'\x01'
    # Out count (VarInt) = 2 (one for OP_RETURN, one for change)
    out_count = b'\x02'
    # Locktime (4 bytes)
    locktime = b'\x00\x00\x00\x00'
    # INPUT:
    #  - Previous TxID (little-endian)
    #  - Previous Vout (4 bytes, little-endian)
    #  - ScriptSig length (varint -> 0 for now, we’ll fill with scriptSig after signing)
    #  - Sequence (4 bytes, e.g. 0xffffffff)
    prev_txid_le = little_endian_hex(prev_txid)
    prev_vout_le = prev_vout.to_bytes(4, 'little')
    sequence = b'\xff\xff\xff\xff'
    # OUTPUT 1: OP_RETURN
    #  - Value (8 bytes, little-endian) = 0 satoshis
    #  - ScriptPubKey = OP_RETURN <message in hex>
    op_return_prefix = b'\x6a'  # OP_RETURN
    msg_hex = message.encode("utf-8")  # raw bytes
    push_len = len(msg_hex)
    # scriptPubKey = OP_RETURN (1 byte) + pushdata length (1 byte) + actual data
    op_return_script = op_return_prefix + push_len.to_bytes(1, 'little') + msg_hex
    op_return_script_len = len(op_return_script)
    value_opreturn = (0).to_bytes(8, 'little')
    op_return_len = op_return_script_len.to_bytes(1, 'little')  # varint (assuming < 0xFD)
    # OUTPUT 2: Change to our address
    # For Testnet P2PKH, version byte is 0x6f, but we’ll reconstruct from pubkey_hash
    # We'll do a standard P2PKH script:
    #   OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
    #   which is: 76 a9 14 <20-byte-script> 88 ac
    p2pkh_prefix = b'\x76\xa9\x14'
    p2pkh_suffix = b'\x88\xac'
    script_pubkey_p2pkh = p2pkh_prefix + pubkey_hash + p2pkh_suffix
    script_pubkey_len = len(script_pubkey_p2pkh).to_bytes(1, 'little')
    value_change = change_value.to_bytes(8, 'little')
    # Put it all together (unsigned for now).
    raw_tx_unsigned = (
        version
        + in_count
        + prev_txid_le
        + prev_vout_le
        + b'\x00'  # scriptSig length placeholder (0 for unsigned)
        + sequence
        + out_count
        + value_opreturn + op_return_len + op_return_script
        + value_change + script_pubkey_len + script_pubkey_p2pkh
        + locktime
    )
    # We need the sighash for signing:
    # SIGHASH_ALL = 0x01
    sighash_all = b'\x01\x00\x00\x00'
    
    # Construct "transaction + scriptPubKey of the input + SIGHASH_ALL"
    # For P2PKH, we put the redeem script = standard scriptPubKey of that input’s address
    # That script is: OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
    redeem_script = p2pkh_prefix + pubkey_hash + p2pkh_suffix
    redeem_script_len = len(redeem_script).to_bytes(1, 'little')
    # Rebuild input section with redeem script for the single input
    raw_tx_for_sig = (
        version
        + in_count
        + prev_txid_le
        + prev_vout_le
        + redeem_script_len + redeem_script
        + sequence
        + out_count
        + value_opreturn + op_return_len + op_return_script
        + value_change + script_pubkey_len + script_pubkey_p2pkh
        + locktime
        + sighash_all
    )
    # Double SHA-256
    h = hash256(raw_tx_for_sig)
    # Sign
    signature = sign_transaction(h, privkey_bytes)
    # Append SIGHASH type 0x01
    signature_plus_hashtype = signature + b'\x01'
    # Final scriptSig = <sig> <pubkey>
    script_sig = (
        len(signature_plus_hashtype).to_bytes(1, 'little') + signature_plus_hashtype
        + len(pubkey_bytes).to_bytes(1, 'little') + pubkey_bytes
    )
    script_sig_len = len(script_sig).to_bytes(1, 'little')
    # Now rebuild the final signed transaction:
    raw_tx_final = (
        version
        + in_count
        + prev_txid_le
        + prev_vout_le
        + script_sig_len + script_sig
        + sequence
        + out_count
        + value_opreturn + op_return_len + op_return_script
        + value_change + script_pubkey_len + script_pubkey_p2pkh
        + locktime
    )
    return binascii.hexlify(raw_tx_final).decode('utf-8')
# ---------------------------------------------------------------------------
# (4) Broadcast via BlockCypher (No requests library)
# ---------------------------------------------------------------------------
def broadcast_tx(hex_tx, blockcypher_token):
    """
    Broadcast a raw transaction hex to BlockCypher using urllib.
    """
    url = "https://api.blockcypher.com/v1/btc/test3/txs/push"
    data = {
        "tx": hex_tx,
        "token": blockcypher_token
    }
    data_bytes = json.dumps(data).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=data_bytes,
        headers={"Content-Type": "application/json"}
    )
    try:
        with urllib.request.urlopen(req) as resp:
            body = resp.read().decode("utf-8")
            js = json.loads(body)
            print("Broadcast success!")
            print("Tx Hash:", js.get("tx", {}).get("hash"))
    except urllib.error.HTTPError as e:
        print("HTTP Error:", e.code)
        err_body = e.read().decode("utf-8")
        print("Error response:", err_body)
    except urllib.error.URLError as e:
        print("URL Error:", e.reason)
# ---------------------------------------------------------------------------
# (5) Example Usage (Main)
# ---------------------------------------------------------------------------
def main():
    # -- You must fill these in manually --
    # 1) Your Testnet WIF private key
    PRIV_WIF = "cNbVaR... (Testnet WIF) ..."  
    # 2) The TXID and output index (vout) you control with the above private key.
    #    This must have enough satoshis to cover your outputs + fee.
    PREV_TXID = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    PREV_VOUT = 0
    PREV_VALUE = 20000  # satoshis in that UTXO
    # 3) OP_RETURN message
    MESSAGE = "Hello, Craig. Leave me alone."
    # 4) BlockCypher token
    BLOCKCYPHER_TOKEN = "8bd4fa2488614e509a677103b88b95fc"
    # 5) Since we’re sending change back to ourselves, we’ll just
    #    reuse the same private key’s address. But in a real scenario,
    #    you’d derive it from the public key. For demonstration,
    #    we assume you’re controlling that same P2PKH output.
    #    (We do not do an address-derivation snippet here.)
    DESTINATION_ADDRESS = "YourTestnetAddressHere"
    print("Creating Raw Transaction...")
    raw_tx_hex = create_raw_transaction(
        priv_wif=PRIV_WIF,
        prev_txid=PREV_TXID,
        prev_vout=PREV_VOUT,
        prev_value=PREV_VALUE,
        destination_address=DESTINATION_ADDRESS,
        message=MESSAGE,
        nettype="test",
    )
    print("Raw Transaction Hex:", raw_tx_hex)
    print("\nBroadcasting...")
    broadcast_tx(raw_tx_hex, BLOCKCYPHER_TOKEN)
if __name__ == "__main__":
    main()
