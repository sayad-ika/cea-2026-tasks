package gchat

import "context"

func CreateSpaceMessage(ctx context.Context, serviceAccountJSON, spaceName string, body []byte) error {
	return CreatePrivateMessage(ctx, serviceAccountJSON, spaceName, "", body)
}
