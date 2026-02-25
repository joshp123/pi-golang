package sdk

import "context"

// Thin RPC mirror layer.
// Each method maps directly to one upstream RPC command.

func (client *Client) Prompt(ctx context.Context, request PromptRequest) error {
	command, err := promptCommand(request)
	if err != nil {
		return err
	}
	_, err = client.send(ctx, command)
	return err
}

func (client *Client) Steer(ctx context.Context, request PromptRequest) error {
	command, err := steerCommand(request)
	if err != nil {
		return err
	}
	_, err = client.send(ctx, command)
	return err
}

func (client *Client) FollowUp(ctx context.Context, request PromptRequest) error {
	command, err := followUpCommand(request)
	if err != nil {
		return err
	}
	_, err = client.send(ctx, command)
	return err
}

func (client *Client) Abort(ctx context.Context) error {
	_, err := client.send(ctx, abortCommand())
	return err
}

func (client *Client) GetState(ctx context.Context) (SessionState, error) {
	response, err := client.send(ctx, getStateCommand())
	if err != nil {
		return SessionState{}, err
	}
	return decodeSessionState(response.Data)
}

func (client *Client) NewSession(ctx context.Context, parentSession string) (bool, error) {
	response, err := client.send(ctx, newSessionCommand(parentSession))
	if err != nil {
		return false, err
	}
	return decodeNewSessionCancelled(response.Data)
}

func (client *Client) Compact(ctx context.Context, customInstructions string) (CompactResult, error) {
	response, err := client.send(ctx, compactCommand(customInstructions))
	if err != nil {
		return CompactResult{}, err
	}
	return decodeCompactResult(response.Data)
}

func (client *SessionClient) ExportHTML(ctx context.Context, outputPath string) (string, error) {
	response, err := client.send(ctx, exportHTMLCommand(outputPath))
	if err != nil {
		return "", err
	}
	return decodeExportPath(response.Data)
}
