package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"

	"github.com/MicahParks/keyfunc"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func ver_jwt(jwtB64 string) (bool, string) {

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
		return false, fmt.Sprintf("Failed to create JWKs from resource at %s.\nError:%s\n", jwksURL, err.Error())
	}

	// Parse the JWT.
	token, err := jwt.Parse(jwtB64, jwks.KeyFunc)
	if err != nil {
		return false, fmt.Sprintf("Failed to parse the JWT.\nToken:%s\nError:%s\n", jwtB64, err.Error())
	}

	// Check if the token is valid.
	if !token.Valid {
		return false, fmt.Sprintf("The token is not valid.")
	}

	return true, "The token is valid."
}
