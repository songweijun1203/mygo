package engine

type SpriteImpl struct {
	sprite Sprite
	name   string
}

func (s *SpriteImpl) GetName() string {
	return s.name
}

func (s *SpriteImpl) SetName(n string) {
	s.name = n
}
