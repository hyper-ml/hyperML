import * as notebookActions from './actions'


export default (
    state = {
        list: [],
        filter: "",
    },
    action) => {
    switch (action.type) {
        case notebookActions.SET_LIST:
            console.log('set_list state:', state)
            let newState = {
                ...state,
                list: action.notebooks,
            }; 
            console.log('newstate:', newState);
            return newState;
        case notebookActions.SET_FILTER:
            return {
                ...state,
                filter: action.filter
            }
        case notebookActions.UPDATE_LIST:
            if (!action.notebook || !action.notebook.ID) {
                return state;
            }

            var list_state = state.list
            var new_list = [] 
            var match_found

            list_state.forEach(function (item, index) {
                if (item.ID === action.notebook.ID) {
                    new_list.push(action.notebook)
                    match_found = true
                } else {
                    new_list.push(item)
                }
            });

            if (!match_found) {
                new_list.push(action.notebook)
            } 

//            console.log('new list:', new_list)
            return {
                ...state,
                list: new_list,
            }            
            
        default:
            console.log('default state:', state)
            return state      
    }
}