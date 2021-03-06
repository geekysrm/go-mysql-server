package sql

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessionConfig(t *testing.T) {
	require := require.New(t)

	sess := NewSession("foo", "bar", 1)

	typ, v := sess.Get("foo")
	require.Equal(Null, typ)
	require.Equal(nil, v)

	sess.Set("foo", Int64, 1)

	typ, v = sess.Get("foo")
	require.Equal(Int64, typ)
	require.Equal(1, v)

	require.Equal(0, len(sess.Warnings()))

	sess.Warn(&Warning{Code: 1})
	sess.Warn(&Warning{Code: 2})
	sess.Warn(&Warning{Code: 3})

	require.Equal(3, len(sess.Warnings()))

	require.Equal(3, sess.Warnings()[0].Code)
	require.Equal(2, sess.Warnings()[1].Code)
	require.Equal(1, sess.Warnings()[2].Code)

}

type testNode struct{}

func (t *testNode) Resolved() bool {
	panic("not implemented")
}

func (t *testNode) TransformUp(func(Node) Node) Node {
	panic("not implemented")
}

func (t *testNode) TransformExpressionsUp(func(Expression) Expression) Node {
	panic("not implemented")
}

func (t *testNode) Schema() Schema {
	panic("not implemented")
}

func (t *testNode) Children() []Node {
	panic("not implemented")
}

func (t *testNode) RowIter(ctx *Context) (RowIter, error) {
	return newTestNodeIterator(ctx), nil
}

type testNodeIterator struct {
	ctx     context.Context
	Counter int
}

func newTestNodeIterator(ctx *Context) RowIter {
	return &testNodeIterator{
		ctx:     ctx,
		Counter: 0,
	}
}

func (t *testNodeIterator) Next() (Row, error) {
	select {
	case <-t.ctx.Done():
		return nil, io.EOF

	default:
		t.Counter++
		return NewRow(true), nil
	}
}

func (t *testNodeIterator) Close() error {
	panic("not implemented")
}

func TestSessionIterator(t *testing.T) {
	require := require.New(t)
	ctx, cancelFunc := context.WithCancel(context.TODO())

	node := &testNode{}
	iter, err := node.RowIter(NewContext(ctx))
	require.NoError(err)

	counter := 0
	for {
		if counter > 5 {
			cancelFunc()
		}

		_, err := iter.Next()

		if counter > 5 {
			require.Equal(io.EOF, err)
			rowIter, ok := iter.(*testNodeIterator)
			require.True(ok)

			require.Equal(counter, rowIter.Counter)
			break
		}

		counter++
	}

	cancelFunc()
}
