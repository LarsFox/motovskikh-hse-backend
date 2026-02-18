package mysql

func (c *Client) Stub() bool {
	return c != nil
}
