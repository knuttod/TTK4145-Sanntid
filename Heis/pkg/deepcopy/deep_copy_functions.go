package deepcopy

import (
	"Heis/pkg/elevator"
	"fmt"
)


func DeepCopy2DSlice[T any](m [][]T) [][]T {
	if m == nil {
		fmt.Println("nil")
		return nil
	}
	deepCopy := make([][]T, len(m))
	for i := range m {
		deepCopy[i] = append([]T{}, m[i]...) // Copy inner slice
	}
	return deepCopy
}

//Since assignedOrders is a map of 2d slice we need to be able to deep copy it to send it between modules
func DeepCopyAssignedOrders(assignedOrders map[string][][]elevator.OrderState) map[string][][]elevator.OrderState {
	deepCopy := make(map[string][][]elevator.OrderState)

	for id, val := range assignedOrders {
		deepCopy[id] = DeepCopy2DSlice(val)
	}
	return deepCopy
}

//Local orders field is a 2d slice, need to deepCopy this
func DeepCopyElevatorStruct(elev elevator.Elevator) elevator.Elevator {
	elev.LocalOrders = DeepCopy2DSlice(elev.LocalOrders)
	return elev
}

func DeepCopyNettworkElevator(elev elevator.NetworkElevator)  elevator.NetworkElevator {
	deepCopy := elevator.NetworkElevator{
		Elevator: DeepCopyElevatorStruct(elev.Elevator),
		AssignedOrders: DeepCopyAssignedOrders(elev.AssignedOrders),
	}

	return deepCopy
}


func DeepCopyElevatorsMap(Elevators map[string] elevator.NetworkElevator) map[string] elevator.NetworkElevator {
	deepCopy := make(map[string]elevator.NetworkElevator)
	
	for key, elev := range Elevators {
		deepCopy[key] = elevator.NetworkElevator{
			Elevator: DeepCopyElevatorStruct(elev.Elevator),
			AssignedOrders: DeepCopyAssignedOrders(elev.AssignedOrders),
		}
	}
	return deepCopy
}