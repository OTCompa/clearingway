package clearingway

import (
	"fmt"
	"strings"

	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/Veraticus/clearingway/internal/fflogs"
)

func legendRoleString(clearedEncounters *Encounters, rankings *fflogs.Rankings) string {
	clears := map[string]*fflogs.Ranking{}

	for _, clearedEncounter := range clearedEncounters.Encounters {
		for _, encounterId := range clearedEncounter.Ids {
			ranking, ok := rankings.Rankings[encounterId]
			if !ok {
				continue
			}
			if !ranking.Cleared() {
				continue
			}

			clears[clearedEncounter.Name] = ranking
		}
	}

	message.Set(language.English, "Cleared the following %d Ultimate fights:\n",
		plural.Selectf(
			1,
			"%d",
			"=1", "Cleared the following one Ultimate fight:\n",
			"=2", "Cleared the following two Ultimate fights:\n",
			"=3", "Cleared the following three Ultimate fights:\n",
			"=4", "Cleared the following four Ultimate fights:\n",
			"other", "Cleared the following Ultimate fights:\n",
		),
	)
	p := message.NewPrinter(language.English)

	clearedString := strings.Builder{}
	clearedString.WriteString(p.Sprintf("Cleared the following %d Ultimate fights:\n", len(clears)))
	for name, ranking := range clears {
		rank := ranking.RanksByTime()[0]
		clearedString.WriteString(
			fmt.Sprintf(
				"     `%v` with `%v` on <t:%v:F> (%v).\n",
				name,
				rank.Job.Abbreviation,
				rank.UnixTime(),
				rank.Report.Url(),
			),
		)
	}

	return strings.TrimSuffix(clearedString.String(), "\n")
}

func LegendRoles() *Roles {
	return &Roles{Roles: []*Role{
		{
			Name: "The Legend", Color: 0x3498db,
			ShouldApply: func(opts *ShouldApplyOpts) (bool, string) {
				clearedEncounters := opts.Encounters.Clears(opts.Rankings)
				if len(clearedEncounters.Encounters) == 1 {
					output := legendRoleString(clearedEncounters, opts.Rankings)
					return true, output
				}

				return false, "Did not clear only one ultimate."
			},
		},
		{
			Name: "The Double Legend", Color: 0x3498db,
			ShouldApply: func(opts *ShouldApplyOpts) (bool, string) {
				clearedEncounters := opts.Encounters.Clears(opts.Rankings)
				if len(clearedEncounters.Encounters) == 2 {
					output := legendRoleString(clearedEncounters, opts.Rankings)
					return true, output
				}

				return false, "Did not clear only two ultimates."
			},
		},
		{
			Name: "The Triple Legend", Color: 0x3498db,
			ShouldApply: func(opts *ShouldApplyOpts) (bool, string) {
				clearedEncounters := opts.Encounters.Clears(opts.Rankings)
				if len(clearedEncounters.Encounters) == 3 {
					output := legendRoleString(clearedEncounters, opts.Rankings)
					return true, output
				}

				return false, "Did not clear only three ultimates."
			},
		},
		{
			Name: "The Quad Legend", Color: 0x3498db,
			ShouldApply: func(opts *ShouldApplyOpts) (bool, string) {
				clearedEncounters := opts.Encounters.Clears(opts.Rankings)
				if len(clearedEncounters.Encounters) == 4 {
					output := legendRoleString(clearedEncounters, opts.Rankings)
					return true, output
				}

				return false, "Did not clear all four ultimates."
			},
		},
	}}
}
