module github.com/zhiqiangxu/evm-lab

go 1.15

require (
	github.com/ethereum/go-ethereum v1.9.25
	github.com/gin-gonic/gin v1.6.3
	github.com/urfave/cli v1.22.5
	github.com/zhiqiangxu/util v0.0.0-20210114025214-5f087283a7a6
)

replace github.com/ethereum/go-ethereum => ../go-ethereum
