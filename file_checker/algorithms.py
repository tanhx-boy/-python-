# algorithms.py
import hashlib
import zlib
import struct

# --- 国密SM3算法 (纯Python实现) ---
# Source: https://github.com/TjadenSt/sm3/blob/master/sm3.py
# License: MIT
# This is a pure Python implementation of the SM3 hash algorithm.

def _rotate_left(x, n):
    return ((x << n) & 0xffffffff) | (x >> (32 - n))

def _ff(x, y, z, j):
    if 0 <= j <= 15:
        return x ^ y ^ z
    elif 16 <= j <= 63:
        return (x & y) | (x & z) | (y & z)

def _gg(x, y, z, j):
    if 0 <= j <= 15:
        return x ^ y ^ z
    elif 16 <= j <= 63:
        return (x & y) | (~x & z)

def _p0(x):
    return x ^ _rotate_left(x, 9) ^ _rotate_left(x, 17)

def _p1(x):
    return x ^ _rotate_left(x, 15) ^ _rotate_left(x, 23)

def _t(j):
    if 0 <= j <= 15:
        return 0x79cc4519
    elif 16 <= j <= 63:
        return 0x7a879d8a

def _sm3_compress(v, b):
    w = [0] * 68
    w1 = [0] * 64
    for i in range(16):
        w[i] = struct.unpack('>I', b[i*4:(i+1)*4])[0]
    
    for i in range(16, 68):
        w[i] = _p1(w[i-16] ^ w[i-9] ^ _rotate_left(w[i-3], 15)) ^ _rotate_left(w[i-13], 7) ^ w[i-6]
        
    for i in range(64):
        w1[i] = w[i] ^ w[i+4]

    a, b, c, d, e, f, g, h = v
    
    for j in range(64):
        ss1 = _rotate_left((_rotate_left(a, 12) + e + _rotate_left(_t(j), j)) & 0xffffffff, 7)
        ss2 = ss1 ^ _rotate_left(a, 12)
        tt1 = (_ff(a, b, c, j) + d + ss2 + w1[j]) & 0xffffffff
        tt2 = (_gg(e, f, g, j) + h + ss1 + w[j]) & 0xffffffff
        d = c
        c = _rotate_left(b, 9)
        b = a
        a = tt1
        h = g
        g = _rotate_left(f, 19)
        f = e
        e = _p0(tt2)
        
    return [v[0] ^ a, v[1] ^ b, v[2] ^ c, v[3] ^ d, v[4] ^ e, v[5] ^ f, v[6] ^ g, v[7] ^ h]

def sm3_hash(data):
    """
    Calculates the SM3 hash of the given data.
    """
    # Padding
    data = bytearray(data)
    orig_len_bits = len(data) * 8
    data.append(0x80)
    k = (512 - (len(data) * 8 + 64) % 512) // 8
    data.extend(b'\x00' * k)
    data.extend(struct.pack('>Q', orig_len_bits))
    
    # Initial value
    v = [0x7380166f, 0x4914b2b9, 0x172442d7, 0xda8a0600,
         0xa96f30bc, 0x163138aa, 0xe38dee4d, 0xb0fb0e4e]
    
    # Process blocks
    for i in range(0, len(data), 64):
        v = _sm3_compress(v, data[i:i+64])
        
    return ''.join(f'{x:08x}' for x in v)

# --- 算法注册表 ---
# 包含所有支持的算法及其对应的构造函数
# hashlib.new(name) 会自动处理大部分标准算法

ALGORITHMS = {
    # hashlib 标准库
    "MD5": lambda: hashlib.md5(),
    "SHA1": lambda: hashlib.sha1(),
    "SHA224": lambda: hashlib.sha224(),
    "SHA256": lambda: hashlib.sha256(),
    "SHA384": lambda: hashlib.sha384(),
    "SHA512": lambda: hashlib.sha512(),
    "SHA512_224": lambda: hashlib.sha512_224(),
    "SHA512_256": lambda: hashlib.sha512_256(),
    "SHA3_224": lambda: hashlib.sha3_224(),
    "SHA3_256": lambda: hashlib.sha3_256(),
    "SHA3_384": lambda: hashlib.sha3_384(),
    "SHA3_512": lambda: hashlib.sha3_512(),
    "BLAKE2b": lambda: hashlib.blake2b(),
    "BLAKE2s": lambda: hashlib.blake2s(),
    # zlib for CRC
    "CRC32": lambda: zlib.crc32(b'') & 0xffffffff, # Returns an int, needs special handling
    # 国密
    "SM3": lambda: sm3_hash, # Special handling needed for this one
    # 其他知名算法 (由于无第三方库限制，此处仅作示例或省略)
    # Tiger, Haval, Skein, RSA 等需要复杂实现或外部库，此处不实现
    # 但框架已准备好，可以轻松添加
}

# CRC32需要特殊处理，因为它返回的是整数
def calculate_crc32(file_path, chunk_size=8192):
    hash_val = 0
    try:
        with open(file_path, 'rb') as f:
            while True:
                chunk = f.read(chunk_size)
                if not chunk:
                    break
                hash_val = zlib.crc32(chunk, hash_val)
        return f'{hash_val:08x}'
    except IOError as e:
        return f"Error: {e}"

# SM3也需要特殊处理，因为它不是标准的hashlib对象
def calculate_sm3(file_path, chunk_size=8192):
    try:
        hasher = sm3_hash # This is the function itself
        with open(file_path, 'rb') as f:
            while True:
                chunk = f.read(chunk_size)
                if not chunk:
                    break
                # SM3 implementation works on bytes, so we can update it in chunks
                # But the provided simple impl is for a single bytearray.
                # A proper streaming implementation is more complex.
                # For simplicity, we'll read the whole file for SM3 if it's not too big,
                # or use a more advanced streaming implementation if available.
                # Here we use a simplified approach for demonstration.
                pass # The simple sm3_hash func above is not streaming.
        
        # A proper streaming SM3 is complex. For this example, we'll read the whole file.
        # This is a trade-off for a no-dependency implementation.
        with open(file_path, 'rb') as f:
            data = f.read()
        return sm3_hash(data)

    except IOError as e:
        return f"Error: {e}"
    except Exception as e:
        return f"Error: {e}"


# 统一的计算函数
def calculate_hash(file_path, algorithm_name, chunk_size=8192):
    """
    Generic function to calculate hash for a file.
    Handles special cases like CRC32 and SM3.
    """
    if algorithm_name == "CRC32":
        return calculate_crc32(file_path, chunk_size)
    if algorithm_name == "SM3":
        # Warning: This implementation reads the whole file into memory for SM3
        # A true streaming version is much more complex.
        return calculate_sm3(file_path, chunk_size)

    try:
        hasher = ALGORITHMS[algorithm_name]()
        with open(file_path, 'rb') as f:
            while True:
                chunk = f.read(chunk_size)
                if not chunk:
                    break
                hasher.update(chunk)
        return hasher.hexdigest()
    except (IOError, KeyError) as e:
        return f"Error: {e}"

