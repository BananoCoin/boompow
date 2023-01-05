package repository

import (
	"fmt"

	"github.com/bananocoin/boompow/apps/server/graph/model"
	"github.com/bananocoin/boompow/apps/server/src/database"
	"github.com/bananocoin/boompow/apps/server/src/models"
	"k8s.io/klog/v2"
)

func UpdateStats(paymentRepo PaymentRepo, workRepo WorkRepo) error {
	// Connected clients
	nConnectedClients, err := database.GetRedisDB().GetNumberConnectedClients()
	if err != nil {
		klog.Infof("Error retrieving connected clients for stats sub %v", err)
		return err
	}
	// Services
	services, err := workRepo.GetServiceStats()
	if err != nil {
		klog.Infof("Error retrieving services for stats sub %v", err)
		return err
	}
	var serviceStats []*model.StatsServiceType
	for _, service := range services {
		serviceStats = append(serviceStats, &model.StatsServiceType{
			Name:     service.ServiceName,
			Website:  service.ServiceWebsite,
			Requests: service.TotalRequests,
		})
	}
	// Top 10
	top10, err := workRepo.GetTopContributors(100)
	if err != nil {
		klog.Infof("Error retrieving # services for stats sub %v", err)
		return err
	}
	var top10Contributors []*model.StatsUserType
	for _, u := range top10 {
		top10Contributors = append(top10Contributors, &model.StatsUserType{
			BanAddress:      u.BanAddress,
			TotalPaidBanano: u.TotalBan,
		})
	}
	// Total paid
	totalPaidBan, err := paymentRepo.GetTotalPaidBanano()
	models.GetStatsInstance().Stats = &model.Stats{ConnectedWorkers: int(nConnectedClients), TotalPaidBanano: fmt.Sprintf("%.2f", totalPaidBan), RegisteredServiceCount: len(services), Top10: top10Contributors, Services: serviceStats}
	return nil
}
