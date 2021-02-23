package client

func GetTestClient() *Client {
	return GetClient("85653da0-9a9e-4262-bc12-f76f2e4e485e", "048a84ac-f107-4b23-9674-1404fafd7fd0")
}

func GetTestClientRateLimiter() *Client {
	return GetClient("85653da0-9a9e-4262-bc12-f76f2e4e485e", "048a84ac-f107-4b23-9674-1404fafd7fd0", RequestInterval(1.2))
}
