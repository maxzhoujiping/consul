package bindingrulelist

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/consul/agent"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/logger"
	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/hashicorp/consul/testrpc"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/require"

	// activate testing idp
	_ "github.com/hashicorp/consul/agent/consul/idp/testing"
)

func TestBindingRuleListCommand_noTabs(t *testing.T) {
	t.Parallel()

	if strings.ContainsRune(New(cli.NewMockUi()).Help(), '\t') {
		t.Fatal("help has tabs")
	}
}

func TestBindingRuleListCommand(t *testing.T) {
	t.Parallel()

	testDir := testutil.TempDir(t, "acl")
	defer os.RemoveAll(testDir)

	a := agent.NewTestAgent(t, t.Name(), `
	primary_datacenter = "dc1"
	acl {
		enabled = true
		tokens {
			master = "root"
		}
	}`)

	a.Agent.LogWriter = logger.NewLogWriter(512)

	defer a.Shutdown()
	testrpc.WaitForLeader(t, a.RPC, "dc1")

	client := a.Client()

	{
		_, _, err := client.ACL().IdentityProviderCreate(
			&api.ACLIdentityProvider{
				Name: "test-1",
				Type: "testing",
			},
			&api.WriteOptions{Token: "root"},
		)
		require.NoError(t, err)

		_, _, err = client.ACL().IdentityProviderCreate(
			&api.ACLIdentityProvider{
				Name: "test-2",
				Type: "testing",
			},
			&api.WriteOptions{Token: "root"},
		)
		require.NoError(t, err)
	}

	createRule := func(t *testing.T, idpName, description string) string {
		rule, _, err := client.ACL().BindingRuleCreate(
			&api.ACLBindingRule{
				IDPName:     idpName,
				Description: description,
				BindType:    api.BindingRuleBindTypeService,
				BindName:    "test-${serviceaccount.name}",
				Selector:    "serviceaccount.namespace==default",
			},
			&api.WriteOptions{Token: "root"},
		)
		require.NoError(t, err)
		return rule.ID
	}

	var ruleIDs []string
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("test-rule-%d", i)

		var idpName string
		if i%2 == 0 {
			idpName = "test-1"
		} else {
			idpName = "test-2"
		}

		id := createRule(t, idpName, name)

		ruleIDs = append(ruleIDs, id)
	}

	t.Run("normal", func(t *testing.T) {
		args := []string{
			"-http-addr=" + a.HTTPAddr(),
			"-token=root",
		}

		ui := cli.NewMockUi()
		cmd := New(ui)

		code := cmd.Run(args)
		require.Equal(t, code, 0)
		require.Empty(t, ui.ErrorWriter.String())
		output := ui.OutputWriter.String()

		for i, v := range ruleIDs {
			require.Contains(t, output, fmt.Sprintf("test-rule-%d", i))
			require.Contains(t, output, v)
		}
	})

	t.Run("filter by idp 1", func(t *testing.T) {
		args := []string{
			"-http-addr=" + a.HTTPAddr(),
			"-token=root",
			"-idp-name=test-1",
		}

		ui := cli.NewMockUi()
		cmd := New(ui)

		code := cmd.Run(args)
		require.Equal(t, code, 0)
		require.Empty(t, ui.ErrorWriter.String())
		output := ui.OutputWriter.String()

		for i, v := range ruleIDs {
			if i%2 == 0 {
				require.Contains(t, output, fmt.Sprintf("test-rule-%d", i))
				require.Contains(t, output, v)
			}
		}
	})

	t.Run("filter by idp 2", func(t *testing.T) {
		args := []string{
			"-http-addr=" + a.HTTPAddr(),
			"-token=root",
			"-idp-name=test-2",
		}

		ui := cli.NewMockUi()
		cmd := New(ui)

		code := cmd.Run(args)
		require.Equal(t, code, 0)
		require.Empty(t, ui.ErrorWriter.String())
		output := ui.OutputWriter.String()

		for i, v := range ruleIDs {
			if i%2 == 1 {
				require.Contains(t, output, fmt.Sprintf("test-rule-%d", i))
				require.Contains(t, output, v)
			}
		}
	})
}
