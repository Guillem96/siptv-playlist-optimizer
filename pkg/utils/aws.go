package utils

import "os"

// IsRunningInLambdaEnv evaluates to true if the current executable is running within an
// AWS lambda function. Internally the function checks if the AWS_LAMBDA_RUNTIME_API is
// present
func IsRunningInLambdaEnv() bool {
	runtime_api, _ := os.LookupEnv("AWS_LAMBDA_RUNTIME_API")
	return runtime_api != ""
}
