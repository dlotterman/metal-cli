// Copyright © 2019 Ali Bazlamit <ali.bazlamit@hotmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package capacity

import (
	"strconv"

	"github.com/packethost/packngo"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// createOrganizationCmd represents the createOrganization command
func (c *Client) Check() *cobra.Command {
	var (
		metro, facility, plan []string
		quantity              int
	)
	var checkCapacityCommand = &cobra.Command{
		Short:   "Validates if a deploy can be fulfilled with the given quantity in any of the given locations and plans",
		Use:     `check {-m [metros,...] | -f [facilities,...]} -P [plans,...] -q [quantity]`,
		Example: `metal capacity check -m sv,ny,da -P c3.large.arm,c3.medium.x86 -q 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var checker func(*packngo.CapacityInput) (*packngo.CapacityInput, *packngo.Response, error)
			var locationField string
			var locationer func(si packngo.ServerInfo) string
			var req = &packngo.CapacityInput{
				Servers: []packngo.ServerInfo{},
			}

			if len(facility) > 0 {
				checker = c.Service.Check
				locationField = "Facility"
				locationer = func(si packngo.ServerInfo) string {
					return si.Facility
				}
				for _, f := range facility {
					for _, p := range plan {
						req.Servers = append(req.Servers, packngo.ServerInfo{
							Facility: f,
							Plan:     p,
							Quantity: quantity},
						)
					}
				}
			} else if len(metro) > 0 {
				checker = c.Service.CheckMetros
				locationField = "Metro"
				locationer = func(si packngo.ServerInfo) string {
					return si.Metro
				}
				for _, m := range metro {
					for _, p := range plan {
						req.Servers = append(req.Servers, packngo.ServerInfo{
							Metro:    m,
							Plan:     p,
							Quantity: quantity},
						)
					}
				}
			} else {
				return errors.New("Either facility or metro should be set")
			}

			availability, _, err := checker(req)
			if err != nil {
				return errors.Wrap(err, "Could not check capacity")
			}

			data := make([][]string, len(availability.Servers))
			for i, s := range availability.Servers {
				data[i] = []string{
					locationer(s),
					s.Plan,
					strconv.Itoa(s.Quantity),
					strconv.FormatBool(s.Available),
				}
			}

			header := []string{locationField, "Plan", "Quantity", "Availability"}
			return c.Out.Output(availability, header, &data)
		},
	}

	checkCapacityCommand.Flags().StringSliceVarP(&metro, "metros", "m", []string{}, "Codes of the metros")
	checkCapacityCommand.Flags().StringSliceVarP(&facility, "facilities", "f", []string{}, "Codes of the facilities")
	checkCapacityCommand.Flags().StringSliceVarP(&plan, "plans", "P", []string{}, "Names of the plans")
	checkCapacityCommand.Flags().IntVarP(&quantity, "quantity", "q", 0, "Number of devices wanted")

	_ = checkCapacityCommand.MarkFlagRequired("plan")
	_ = checkCapacityCommand.MarkFlagRequired("quantity")
	return checkCapacityCommand
}
