package config

import "github.com/liyanze888/funny-core/fn_factory"

const (
	ReadDiffusion  = 1 //读扩散
	WriteDiffusion = 2 //写扩散
)

func init() {
	fn_factory.BeanFactory.RegisterBean(NewGlobalConfig())
}

type GlobalConfig struct {
	Diffusion int //	1 读扩散  2写扩散
}

func NewGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Diffusion: WriteDiffusion,
	}
}
