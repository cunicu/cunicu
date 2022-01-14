package net

func Filtered(p *NetworkParams) (*Network, error) {
	n, err := Simple(p)
	if err != nil {
		return nil, err
	}

	for _, n := range n.Nodes {
		n.Host.Run("nft")
	}

	return n, nil
}
