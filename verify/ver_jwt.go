package verify

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"time"

	// "github.com/golang-jwt/jwt"

	"github.com/dgrijalva/jwt-go"

	"github.com/MicahParks/keyfunc"
)

type person struct {
	name string
	age  int
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func VerifyJWT(jwtB64 string) (jwt.MapClaims, error) {

	bob := person{
		name: "MacDonald, Bob",
		age:  8,
	}

	// Get the JWKs URL from your AWS region and userPoolId.
	//
	// See the AWS docs here:
	// https://docs.aws.amazon.com/cognito/latest/developerguide/amazon-cognito-user-pools-using-tokens-verifying-a-jwt.html
	regionID := getEnv("AWS_REGION", "eu-west-2")
	userPoolID := getEnv("COGNITO_USERPOOL_ID", "")
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", regionID, userPoolID)

	// Create the keyfunc options. Use an error handler that logs. Refresh the JWKs when a JWT signed by an unknown KID
	// is found or at the specified interval. Rate limit these refreshes. Timeout the initial JWKs refresh request after
	// 10 seconds. This timeout is also used to create the initial context.Context for keyfunc.Get.
	refreshInterval := time.Hour
	refreshRateLimit := time.Minute * 5
	refreshTimeout := time.Second * 10
	refreshUnknownKID := true
	options := keyfunc.Options{
		RefreshErrorHandler: func(err error) {
			log.Printf("There was an error with the jwt.KeyFunc\nError:%s\n", err.Error())
		},
		RefreshInterval:   &refreshInterval,
		RefreshRateLimit:  &refreshRateLimit,
		RefreshTimeout:    &refreshTimeout,
		RefreshUnknownKID: &refreshUnknownKID,
	}

	// Create the JWKs from the resource at the given URL.
	jwks, err := keyfunc.Get(jwksURL, options)
	if err != nil {
		return nil, fmt.Errorf("Failed to create JWKs from resource at %s.\nError:%s\n", jwksURL, err.Error())
	}

	// Parse the JWT.
	//token, err := jwt.Parse(jwtB64, jwks.KeyFunc)

	token, err := jwt.Parse(jwtB64, func(t *jwt.Token) (interface{}, error) {
		jjj := jwks.KeyFunc
		claims := t.Claims.(jwt.MapClaims)
		log.Println(claims)
		log.Println(claims["iss"])
		log.Println(claims["name"])
		log.Println(claims["email"])
		log.Println(jjj)
		log.Println(bob)
		return bob, nil
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to parse the JWT.\nToken:%s\nError:%s\n", jwtB64, err.Error())
	}

	// Check if the token is valid.
	if !token.Valid {
		return nil, fmt.Errorf("The token is not valid.")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("Token claim is not valid!")
	}
}
