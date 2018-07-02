package ravendb

type BatchOperation struct {
	_session              *InMemoryDocumentSessionOperations
	_entities             []Object
	_sessionCommandsCount int
}

func NewBatchOperation(session *InMemoryDocumentSessionOperations) *BatchOperation {
	return &BatchOperation{
		_session: session,
	}
}

func (b *BatchOperation) createRequest() *BatchCommand {
	result := b._session.prepareForSaveChanges()

	b._sessionCommandsCount = len(result.getSessionCommands())
	result.sessionCommands = append(result.sessionCommands, result.getDeferredCommands()...)
	if len(result.getSessionCommands()) == 0 {
		return nil
	}

	// TODO: should propagate an error
	b._session.incrementRequestCount()

	b._entities = result.getEntities()

	return NewBatchCommandWithOptions(b._session.getConventions(), result.getSessionCommands(), result.getOptions())
}

func (b *BatchOperation) setResult(result ArrayNode) {
	if len(result) == 0 {
		// TODO: throwOnNullResults()
		return
	}
	for i := 0; i < b._sessionCommandsCount; i++ {
		batchResult := result[i]
		if batchResult == nil {
			return
			//TODO: throw new IllegalArgumentException();
		}
		typ, _ := jsonGetAsText(batchResult, "Type")
		if typ != "PUT" {
			continue
		}
		entity := b._entities[i]
		documentInfo := b._session.documentsByEntity[entity]
		if documentInfo == nil {
			continue
		}
		changeVector := jsonGetAsTextPointer(batchResult, Constants_Documents_Metadata_CHANGE_VECTOR)
		if changeVector == nil {
			return
			//TODO: throw new IllegalStateException("PUT response is invalid. @change-vector is missing on " + documentInfo.getId());
		}
		id, _ := jsonGetAsText(batchResult, Constants_Documents_Metadata_ID)
		if id == "" {
			return
			//TODO: throw new IllegalStateException("PUT response is invalid. @id is missing on " + documentInfo.getId());
		}

		for propertyName, v := range batchResult {
			if propertyName == "Type" {
				continue
			}

			meta := documentInfo.getMetadata()
			meta[propertyName] = v
		}

		documentInfo.setId(id)
		documentInfo.setChangeVector(changeVector)
		doc := documentInfo.getDocument()
		doc[Constants_Documents_Metadata_KEY] = documentInfo.getMetadata()
		documentInfo.setMetadataInstance(nil)

		b._session.documentsById.add(documentInfo)
		b._session.getGenerateEntityIdOnTheClient().trySetIdentity(entity, id)

		afterSaveChangesEventArgs := NewAfterSaveChangesEventArgs(b._session, documentInfo.getId(), documentInfo.getEntity())
		b._session.onAfterSaveChangesInvoke(afterSaveChangesEventArgs)
	}
}
