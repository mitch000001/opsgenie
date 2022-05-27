package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/kr/pretty"
	"github.com/opsgenie/opsgenie-go-sdk-v2/schedule"
	"github.com/spf13/cobra"
)

var schedulesCmd = &cobra.Command{
	Use:   "schedules",
	Short: "Get all schedules",
	RunE: func(cmd *cobra.Command, args []string) error {
		schedulesAPI, cliErr := schedule.NewClient(opsgenieConfig)
		if cliErr != nil {
			return fmt.Errorf("error creating ScheduleV2 client: %w", cliErr)
		}
		return getSchedules(context.Background(), schedulesAPI, expandRotations)
	},
}

var expandRotations bool

var rotationCmd = &cobra.Command{
	Use:   "rotations scheduleName",
	Short: "Prints the rotation for a given schedule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rotaAPI, cliErr := schedule.NewClient(opsgenieConfig)
		if cliErr != nil {
			return fmt.Errorf("error creating ScheduleV2 client: %w", cliErr)
		}
		return getRotations(context.Background(), rotaAPI, args[0])
	},
}

var schedTimelineCmd = &cobra.Command{
	Use:   "timeline scheduleName",
	Short: "Prints the timeline for a given schedule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		schedulesAPI, cliErr := schedule.NewClient(opsgenieConfig)
		if cliErr != nil {
			return fmt.Errorf("error creating Schedule client: %w", cliErr)
		}
		startDate, err := time.Parse("2006-01-02", timelineStartDate)
		if err != nil {
			return fmt.Errorf("error parsing timeline start date: %w", err)
		}
		timeline, err := getScheduleTimeline(context.Background(), schedulesAPI, args[0], startDate, scheduleTimelineInterval)
		if err != nil {
			return fmt.Errorf("error getting schedule timeline: %w", err)
		}
		pretty.Println(timeline)
		return nil
	},
}

var timelineStartDate string
var scheduleTimelineInterval scheduleInterval = scheduleInterval{unit: schedule.Days, value: 14}

func init() {
	rootCmd.AddCommand(schedulesCmd)
	rootCmd.AddCommand(rotationCmd)
	// schedulesCmd.Flags().String("user", "", "The user for which the schedules should be fetched")
	schedulesCmd.Flags().BoolVar(&expandRotations, "expand-rotations", false, "When enabled the rotations will also be printed")
	schedulesCmd.AddCommand(schedTimelineCmd)
	schedTimelineCmd.Flags().StringVar(&timelineStartDate, "start-date", time.Now().Format("2006-01-02"), "The start date from which to relate the timeline from. Defaults to today.")
	schedTimelineCmd.Flags().Var(&scheduleTimelineInterval, "interval", "The interval to fetch the timeline from starting at start-date. Defaults to 14 days.")
}

func getSchedules(ctx context.Context, schedAPI *schedule.Client, expandRotations bool) error {
	resp, err := schedAPI.List(ctx, &schedule.ListRequest{
		Expand: &expandRotations,
	})
	if err != nil {
		return fmt.Errorf("error getting schedules: %w", err)
	}
	pretty.Println(resp.Schedule)
	return nil
}

func getRotations(ctx context.Context, rotaApi *schedule.Client, scheduleName string) error {
	resp, err := rotaApi.ListRotations(ctx, &schedule.ListRotationsRequest{
		ScheduleIdentifierType:  schedule.Name,
		ScheduleIdentifierValue: scheduleName,
	})
	if err != nil {
		return fmt.Errorf("error getting schedule rotations: %w", err)
	}
	for _, rot := range resp.Rotations {
		fmt.Printf("%+#v\n", rot)
	}
	return nil
}

// func getRotation(ctx context.Context, rotaApi *schedule.Client, schedule string, rotation string) error {
// 	return nil
// }

func getScheduleTimeline(ctx context.Context, schedAPI *schedule.Client, scheduleName string, startDate time.Time, scheduleTimelineInterval scheduleInterval) (map[identifier][]period, error) {
	timeline, err := schedAPI.GetTimeline(ctx, &schedule.GetTimelineRequest{
		IdentifierType:  schedule.Name,
		IdentifierValue: scheduleName,
		IntervalUnit:    scheduleTimelineInterval.unit,
		Interval:        scheduleTimelineInterval.value,
		Date:            &startDate,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting timeline for schedule %q: %w", scheduleName, err)
	}
	rotationPeriods := map[identifier][]period{}
	for _, rot := range timeline.FinalTimeline.Rotations {
		id := identifier{Name: rot.Name, ID: rot.Id}
		rotationPeriods[id] = collectPeriods(rot)
	}
	return rotationPeriods, nil
}

func collectPeriods(rot schedule.TimelineRotation) []period {
	var periods []period
	for i, p := range rot.Periods {
		current := period{
			StartDate: p.StartDate,
			EndDate:   p.EndDate,
			OnCallee:  identifier{Name: p.Recipient.Name, ID: p.Recipient.Id},
		}
		if i == 0 {
			periods = append(periods, current)
			continue
		}
		prev := periods[len(periods)-1]
		if prev.OnCallee == current.OnCallee {
			prev.EndDate = current.EndDate
			periods[len(periods)-1] = prev
			continue
		}
		periods = append(periods, current)
	}
	return periods
}

type identifier struct {
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

type period struct {
	StartDate time.Time  `yaml:"startDate"`
	EndDate   time.Time  `yaml:"endDate"`
	OnCallee  identifier `yaml:"onCallee"`
}

var scheduleIntervalRegexp = regexp.MustCompile(`(\d+)(days|weeks|months)`)

type scheduleInterval struct {
	unit  schedule.Unit
	value int
}

func (s *scheduleInterval) String() string {
	return fmt.Sprintf("%d%s", s.value, s.unit)
}

func (s *scheduleInterval) Set(value string) error {
	matches := scheduleIntervalRegexp.FindStringSubmatch(value)
	if len(matches) != 3 {
		return fmt.Errorf("value does not comply with regexp %q", scheduleIntervalRegexp.String())
	}
	val, err := strconv.Atoi(matches[1])
	if err != nil {
		return fmt.Errorf("error parsing interval value: %w", err)
	}
	s.value = val
	s.unit = schedule.Unit(matches[2])
	return nil
}

func (s *scheduleInterval) Type() string {
	return "string"
}
