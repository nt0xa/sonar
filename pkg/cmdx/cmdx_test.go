package cmdx_test

import (
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/pkg/cmdx"
)

// tagWrapper appends tag to *trace so wrapper ordering can be observed in tests.
func tagWrapper(trace *[]string, tag string) cmdx.Wrapper {
	return func(next cmdx.RunFunc) cmdx.RunFunc {
		return func(cmd *cobra.Command, args []string) error {
			*trace = append(*trace, tag)
			return next(cmd, args)
		}
	}
}

// recordRun returns a RunFunc that flips *ran to true.
func recordRun(ran *bool) cmdx.RunFunc {
	return func(_ *cobra.Command, _ []string) error {
		*ran = true
		return nil
	}
}

// exec runs the tree with the given args.
func exec(t *testing.T, root *cmdx.Command, args ...string) error {
	t.Helper()
	c := root.Cobra()
	c.SetArgs(args)
	c.SetOut(io.Discard)
	c.SetErr(io.Discard)
	c.SilenceUsage = true
	c.SilenceErrors = true
	return c.Execute()
}

func TestNew_Add(t *testing.T) {
	root := cmdx.New("root", "Root command")
	require.NotNil(t, root)

	var ran bool
	root.Add("foo", "Foo", recordRun(&ran))

	require.NoError(t, exec(t, root, "foo"))
	assert.True(t, ran)
}

func TestAddCommand_Prebuilt(t *testing.T) {
	// A prebuilt cobra command with its own RunE and a flag is ingested, and the
	// accumulated wrappers wrap that existing RunE.
	var trace []string
	root := cmdx.New("root", "Root command", tagWrapper(&trace, "a"))

	var ran bool
	var flag string
	sub := &cobra.Command{
		Use: "foo",
		RunE: func(_ *cobra.Command, _ []string) error {
			ran = true
			return nil
		},
	}
	sub.Flags().StringVar(&flag, "name", "", "name flag")

	root.AddCommand(sub)

	require.NoError(t, exec(t, root, "foo", "--name", "bar"))
	assert.True(t, ran)
	assert.Equal(t, "bar", flag)
	assert.Equal(t, []string{"a"}, trace)
}

func TestNew_DefaultGroup(t *testing.T) {
	root := cmdx.New("root", "Root command")

	// The "main" group is registered on the root.
	groups := root.Cobra().Groups()
	require.Len(t, groups, 1)
	assert.Equal(t, "main", groups[0].ID)

	// Top-level Add and Group tag their commands into it.
	root.Add("foo", "Foo", recordRun(new(bool)))
	root.Group("bar", "Bar", func(_ *cmdx.Command) {})

	byName := map[string]*cobra.Command{}
	for _, c := range root.Cobra().Commands() {
		byName[c.Name()] = c
	}
	assert.Equal(t, "main", byName["foo"].GroupID)
	assert.Equal(t, "main", byName["bar"].GroupID)
}

func TestWrap_Order(t *testing.T) {
	// First registered wrapper is the outermost handler.
	var trace []string
	root := cmdx.New("root", "Root command", tagWrapper(&trace, "first"))
	root.Wrap(tagWrapper(&trace, "second"))
	root.Add("foo", "Foo", recordRun(new(bool)))

	require.NoError(t, exec(t, root, "foo"))
	assert.Equal(t, []string{"first", "second"}, trace)
}

func TestWrap_ShortCircuit(t *testing.T) {
	sentinel := assert.AnError
	stop := func(next cmdx.RunFunc) cmdx.RunFunc {
		return func(_ *cobra.Command, _ []string) error {
			return sentinel // never calls next
		}
	}

	root := cmdx.New("root", "Root command", stop)

	var ran bool
	root.Add("foo", "Foo", recordRun(&ran))

	err := exec(t, root, "foo")
	assert.ErrorIs(t, err, sentinel)
	assert.False(t, ran)
}

func TestGroup_Nested(t *testing.T) {
	root := cmdx.New("root", "Root command")

	var ran bool
	root.Group("sub", "Sub", func(c *cmdx.Command) {
		c.Add("leaf", "Leaf", recordRun(&ran))
	})

	require.NoError(t, exec(t, root, "sub", "leaf"))
	assert.True(t, ran)

	// The container alone is non-runnable and prints help without error.
	assert.NoError(t, exec(t, root, "sub"))
}

func TestGroup_InheritsWrappers(t *testing.T) {
	var trace []string
	root := cmdx.New("root", "Root command", tagWrapper(&trace, "base"))

	root.Group("sub", "Sub", func(c *cmdx.Command) {
		c.Wrap(tagWrapper(&trace, "group"))
		c.Add("leaf", "Leaf", recordRun(new(bool)))
	})

	require.NoError(t, exec(t, root, "sub", "leaf"))
	assert.Equal(t, []string{"base", "group"}, trace)
}

func TestGroup_IsolatesWrappers(t *testing.T) {
	var trace []string
	root := cmdx.New("root", "Root command", tagWrapper(&trace, "base"))

	root.Group("sub", "Sub", func(c *cmdx.Command) {
		c.Wrap(tagWrapper(&trace, "group"))
		c.Add("leaf", "Leaf", recordRun(new(bool)))
	})

	// A wrapper added inside the group must not leak to a root sibling.
	root.Add("outer", "Outer", recordRun(new(bool)))

	require.NoError(t, exec(t, root, "sub", "leaf"))
	assert.Equal(t, []string{"base", "group"}, trace)

	trace = nil
	require.NoError(t, exec(t, root, "outer"))
	assert.Equal(t, []string{"base"}, trace)
}

func TestAdd_Opts(t *testing.T) {
	root := cmdx.New("root", "Root command")

	var flag string
	root.Add("foo", "Foo",
		func(_ *cobra.Command, _ []string) error { return nil },
		func(cmd *cobra.Command) {
			cmd.Flags().StringVar(&flag, "name", "", "name flag")
		},
	)

	require.NoError(t, exec(t, root, "foo", "--name", "baz"))
	assert.Equal(t, "baz", flag)
}
