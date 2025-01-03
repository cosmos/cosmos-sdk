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
