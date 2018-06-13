package ravendb

import "time"

type Operation struct {
	_requestExecutor *RequestExecutor
	//TBD private readonly Func<IDatabaseChanges> _changes;
	_conventions *DocumentConventions
	_id          int
}

func (o *Operation) getId() int {
	return o._id
}

func NewOperation(requestExecutor *RequestExecutor, changes *IDatabaseChanges, conventions *DocumentConventions, id int) *Operation {
	return &Operation{
		_requestExecutor: requestExecutor,
		//TBD _changes = changes;
		_conventions: conventions,
		_id:          id,
	}
}

func (o *Operation) fetchOperationsStatus() (ObjectNode, error) {
	command := o.getOperationStateCommand(o._conventions, o._id)
	err := o._requestExecutor.executeCommand(command)
	if err != nil {
		return nil, err
	}
	return command.Result, nil
}

func (o *Operation) getOperationStateCommand(conventions *DocumentConventions, id int) *GetOperationStateCommand {
	return NewGetOperationStateCommand(o._conventions, o._id)
}

func (o *Operation) waitForCompletion() error {
	for {
		status, err := o.fetchOperationsStatus()
		if err != nil {
			return err
		}

		operationStatus := jsonGetAsText(status, "Status")
		switch operationStatus {
		case "Completed":
			return nil
		case "Cancelled":
			return NewOperationCancelledException("")
		case "Faulted":
			panicIf(true, "NYI")
			/*
				result := status["Result"]

				OperationExceptionResult exceptionResult = JsonExtensions.getDefaultMapper().convertValue(result, OperationExceptionResult.class);

				throw ExceptionDispatcher.get(exceptionResult.getMessage(), exceptionResult.getError(), exceptionResult.getType(), exceptionResult.getStatusCode());
			*/
		}

		time.Sleep(500 * time.Millisecond)
	}
}
