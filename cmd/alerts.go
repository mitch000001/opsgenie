package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(alertsCmd)
	alertsCmd.Flags().StringVar(&alertAcknowledgedByUser, "acknowledged-by", "", "Lists only the alerts acknowledged by the given user")
	alertsCmd.Flags().Var(&dateFlag{&alertListStartDate}, "start-date", "Lists only the alerts starting from that date")
	alertsCmd.Flags().Var(&dateFlag{&alertListEndDate}, "end-date", "Lists only the alerts ending at that date")
}

var alertAcknowledgedByUser string
var alertListStartDate time.Time
var alertListEndDate time.Time

var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Lists alerts",
	RunE: func(cmd *cobra.Command, args []string) error {
		alertAPI, cliErr := alert.NewClient(opsgenieConfig)
		if cliErr != nil {
			return fmt.Errorf("error creating Alert client: %w", cliErr)
		}
		alerts, err := listAlerts(context.Background(), alertAPI)
		if err != nil {
			return fmt.Errorf("error listing alerts: %w", err)
		}
		pretty.Println(alerts)
		return nil
	},
}

func listAlerts(ctx context.Context, client *alert.Client) ([]alert.Alert, error) {
	queryElements := []string{}
	if alertAcknowledgedByUser != "" {
		queryElements = append(
			queryElements,
			fmt.Sprintf("acknowledgedBy:%s", alertAcknowledgedByUser),
		)
	}
	if !alertListStartDate.IsZero() {
		queryElements = append(
			queryElements,
			fmt.Sprintf("createdAt:%s", alertListStartDate.Format("02-01-2006")),
		)
	}
	if !alertListEndDate.IsZero() {
		queryElements = append(
			queryElements,
			fmt.Sprintf("createdAt:%s", alertListEndDate.Format("02-01-2006")),
		)
	}
	resp, err := client.List(ctx, &alert.ListAlertRequest{
		Sort:  alert.CreatedAt,
		Query: strings.Join(queryElements, " AND "),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting alerts: %w", err)
	}
	return resp.Alerts, nil
}

type dateFlag struct {
	date *time.Time
}

func (d *dateFlag) String() string {
	return d.date.Format("2006-01-02")
}

func (d *dateFlag) Set(value string) error {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return fmt.Errorf("error parsing date: %w", err)
	}
	if d.date == nil {
		d.date = new(time.Time)
	}
	*d.date = date
	return nil
}

func (d *dateFlag) Type() string {
	return "string"
}
