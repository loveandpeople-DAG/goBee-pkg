package dashboard

import (
	"time"

	"github.com/loveandpeople-DAG/goHive/daemon"
	"github.com/loveandpeople-DAG/goHive/events"

	"github.com/loveandpeople-DAG/goBee/pkg/model/hornet"
	"github.com/loveandpeople-DAG/goBee/pkg/model/milestone"
	tanglemodel "github.com/loveandpeople-DAG/goBee/pkg/model/tangle"
	"github.com/loveandpeople-DAG/goBee/pkg/shutdown"
	"github.com/loveandpeople-DAG/goBee/plugins/tangle"
)

func runLiveFeed() {

	newTxZeroValueRateLimiter := time.NewTicker(time.Second / 10)
	newTxValueRateLimiter := time.NewTicker(time.Second / 20)

	onReceivedNewTransaction := events.NewClosure(func(cachedTx *tanglemodel.CachedTransaction, latestMilestoneIndex milestone.Index, latestSolidMilestoneIndex milestone.Index) {
		cachedTx.ConsumeTransaction(func(tx *hornet.Transaction) {
			if !tanglemodel.IsNodeSyncedWithThreshold() {
				return
			}

			if tx.Tx.Value == 0 {
				select {
				case <-newTxZeroValueRateLimiter.C:
					hub.BroadcastMsg(&Msg{Type: MsgTypeTxZeroValue, Data: &LivefeedTransaction{Hash: tx.Tx.Hash, Value: tx.Tx.Value}})
				default:
				}
			} else {
				select {
				case <-newTxValueRateLimiter.C:
					hub.BroadcastMsg(&Msg{Type: MsgTypeTxValue, Data: &LivefeedTransaction{Hash: tx.Tx.Hash, Value: tx.Tx.Value}})
				default:
				}
			}
		})
	})

	onLatestMilestoneIndexChanged := events.NewClosure(func(msIndex milestone.Index) {
		if msTailTxHash := getMilestoneTailHash(msIndex); msTailTxHash != nil {
			hub.BroadcastMsg(&Msg{Type: MsgTypeMs, Data: &LivefeedMilestone{Hash: msTailTxHash.Trytes(), Index: msIndex}})
		}
	})

	daemon.BackgroundWorker("Dashboard[TxUpdater]", func(shutdownSignal <-chan struct{}) {
		tangle.Events.ReceivedNewTransaction.Attach(onReceivedNewTransaction)
		defer tangle.Events.ReceivedNewTransaction.Detach(onReceivedNewTransaction)
		tangle.Events.LatestMilestoneIndexChanged.Attach(onLatestMilestoneIndexChanged)
		defer tangle.Events.LatestMilestoneIndexChanged.Detach(onLatestMilestoneIndexChanged)

		<-shutdownSignal

		log.Info("Stopping Dashboard[TxUpdater] ...")
		newTxZeroValueRateLimiter.Stop()
		newTxValueRateLimiter.Stop()
		log.Info("Stopping Dashboard[TxUpdater] ... done")
	}, shutdown.PriorityDashboard)
}
