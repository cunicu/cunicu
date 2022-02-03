package pb

func (s *SignalingMessage) SessionDescription() *SessionDescription {
	switch s := s.Msg.(type) {
	case *SignalingMessage_Answer:
		return s.Answer
	case *SignalingMessage_Offer:
		return s.Offer
	case *SignalingMessage_Candidates:
		return s.Candidates
	default:
		return nil
	}
}
