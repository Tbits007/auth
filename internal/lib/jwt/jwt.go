package jwt

import (
	"time"

	"github.com/Tbits007/auth/internal/domain/models/userModel"
	"github.com/golang-jwt/jwt/v5"
)


func NewToken(
	user userModel.User,
	duration time.Duration, 
	secretKey string,
) (string, error) {  
    token := jwt.New(jwt.SigningMethodHS256)  

    claims := token.Claims.(jwt.MapClaims)  
    claims["uuid"] = user.ID  
    claims["email"] = user.Email  
    claims["exp"] = time.Now().Add(duration).Unix()   

    tokenString, err := token.SignedString([]byte(secretKey))  
    if err != nil {  
       return "", err  
    }  

    return tokenString, nil  
}