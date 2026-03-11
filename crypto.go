package weworkaibotsdk

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// ref: https://github.com/WecomTeam/aibot-node-sdk/blob/main/src/crypto.ts
// 加解密工具模块
// 提供文件加解密相关的功能函数

// DecryptFile 使用 AES-256-CBC 解密文件
//
// 参数:
//   - encryptedBuffer: 加密的文件数据
//   - aesKey: Base64 编码的 AES-256 密钥
//
// 返回:
//   - 解密后的文件数据
//   - 错误信息（如果有）
func DecryptFile(encryptedBuffer []byte, aesKey string) ([]byte, error) {
	// 参数验证
	if len(encryptedBuffer) == 0 {
		return nil, errors.New("decryptFile: encryptedBuffer is empty or not provided")
	}

	if aesKey == "" {
		return nil, errors.New("decryptFile: aesKey must be a non-empty string")
	}

	// 将 Base64 编码的 aesKey 解码为字节数组，目前消息中的aesKey大部分没有padding
	key, err := base64.RawStdEncoding.DecodeString(strings.TrimRight(aesKey, "="))
	if err != nil {
		return nil, fmt.Errorf("decryptFile: failed to decode base64 aesKey - %w", err)
	}

	// IV 取 aesKey 解码后的前 16 字节
	if len(key) < 16 {
		return nil, errors.New("decryptFile: decoded key is too short, must be at least 16 bytes")
	}
	iv := key[:16]

	// 创建 AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("decryptFile: failed to create AES cipher - %w", err)
	}

	// 验证加密数据长度是否是块大小的倍数
	if len(encryptedBuffer)%aes.BlockSize != 0 {
		return nil, errors.New("decryptFile: encrypted data is not a multiple of block size")
	}

	// 创建 CBC 模式解密器
	mode := cipher.NewCBCDecrypter(block, iv)

	// 解密数据（原地解密）
	decrypted := make([]byte, len(encryptedBuffer))
	mode.CryptBlocks(decrypted, encryptedBuffer)

	// 手动去除 PKCS#7 填充（支持 32 字节 block）
	if len(decrypted) == 0 {
		return nil, errors.New("decryptFile: decrypted data is empty")
	}

	padLen := int(decrypted[len(decrypted)-1])
	if padLen < 1 || padLen > 32 || padLen > len(decrypted) {
		return nil, fmt.Errorf("decryptFile: invalid PKCS#7 padding value: %d", padLen)
	}

	// 验证所有 padding 字节是否一致
	for i := len(decrypted) - padLen; i < len(decrypted); i++ {
		if decrypted[i] != byte(padLen) {
			return nil, errors.New("decryptFile: invalid PKCS#7 padding: padding bytes mismatch")
		}
	}

	// 返回去除填充后的数据
	return decrypted[:len(decrypted)-padLen], nil
}
