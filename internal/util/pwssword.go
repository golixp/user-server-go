package util

import "golang.org/x/crypto/bcrypt"

// HashAndSaltPassword 使用 bcrypt 对明文密码进行加盐哈希处理。
// 返回哈希字符串和错误（如果发生）。
//
// 示例：
//
//	hashed, err := util.HashAndSaltPassword("mySecret123")
//	if err != nil { ... }
//	fmt.Println("Hashed password:", hashed)
func HashAndSaltPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword 验证明文密码与哈希密码是否匹配。
// 返回 true 表示密码匹配，false 表示不匹配或格式非法。
//
// 示例：
//
//	match := util.VerifyPassword("mySecret123", hashedPassword)
//	if match {
//	    fmt.Println("Password is correct")
//	} else {
//	    fmt.Println("Password is incorrect")
//	}
func VerifyPassword(password, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}
