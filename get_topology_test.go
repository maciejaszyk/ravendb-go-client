package ravendb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTopologyTest_canGetTopology(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	command := NewGetDatabaseTopologyCommand()
	err = store.GetRequestExecutor().executeCommand(command)
	assert.NoError(t, err)
	result := command.Result
	assert.NotNil(t, result)

	assert.NotEqual(t, result.getEtag(), "")
	assert.Equal(t, len(result.getNodes()), 1)
	serverNode := result.getNodes()[0]
	assert.Equal(t, serverNode.getUrl(), store.getUrls()[0])
	assert.Equal(t, serverNode.getDatabase(), store.getDatabase())
	assert.Equal(t, serverNode.getClusterTag(), "A")
	assert.Equal(t, serverNode.getServerRole(), ServerNode_Role_MEMBER)
}

func TestGetTopology(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer func() {
		r := recover()
		destroyDriver()
		if r != nil {
			panic(r)
		}
	}()

	getTopologyTest_canGetTopology(t)
}
