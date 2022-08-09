package launcher

import (
	"fmt"
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/inter"
)

func checkEvm(ctx *cli.Context) error {
	if len(ctx.Args()) != 0 {
		utils.Fatalf("This command doesn't require an argument.")
	}

	cfg := makeAllConfigs(ctx)

	rawDbs := makeDirectDBsProducer(cfg)

	prefixes := []byte{'E','F','G','H','I','J','Q','R','S','T','U','e'}
	for _, prefix := range prefixes {
		db, err := rawDbs.OpenDB("gossip/"+string([]byte{prefix}))
		if err != nil {
			utils.Fatalf("unable to open gossip; %s", err)
		}

		it := db.NewIterator([]byte{}, nil)
		if it.Next() {
			fmt.Printf("Prefix %c is NOT EMPTY\n", prefix)
		} else {
			fmt.Printf("Prefix %c is EMPTY\n", prefix)
		}
		it.Release()
		db.Close()
	}
	utils.Fatalf("Prefixes checking done")

	gdb := makeGossipStore(rawDbs, cfg)
	defer gdb.Close()
	evms := gdb.EvmStore()

	start, reported := time.Now(), time.Now()

	var prevPoint idx.Block
	checkBlocks := func(stateOK func(root common.Hash) (bool, error)) {
		var (
			lastIdx            = gdb.GetLatestBlockIndex()
			prevPointRootExist bool
		)
		gdb.ForEachBlock(func(index idx.Block, block *inter.Block) {
			found, err := stateOK(common.Hash(block.Root))
			if found != prevPointRootExist {
				if index > 0 && found {
					log.Warn("EVM history is pruned", "fromBlock", prevPoint, "toBlock", index-1)
				}
				prevPointRootExist = found
				prevPoint = index
			}
			if index == lastIdx && !found {
				log.Crit("State trie for the latest block is not found", "block", index)
			}
			if !found {
				return
			}
			if err != nil {
				log.Crit("State trie error", "err", err, "block", index)
			}
			if time.Since(reported) >= statsReportLimit {
				log.Info("Checking presence of every node", "last", index, "pruned", !prevPointRootExist, "elapsed", common.PrettyDuration(time.Since(start)))
				reported = time.Now()
			}
		})
	}

	if err := evms.CheckEvm(checkBlocks); err != nil {
		return err
	}
	log.Info("EVM storage is verified", "last", prevPoint, "elapsed", common.PrettyDuration(time.Since(start)))
	return nil
}
