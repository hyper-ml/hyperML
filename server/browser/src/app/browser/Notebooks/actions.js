import api from '../../api'


export const SET_LIST = "notebooks/SET_LIST"
export const SET_FILTER = "notebooks/SET_FILTER"
export const UPDATE_LIST = "notebooks/UPDATE_LIST"

export const fetchNotebooks = () => {
    return function(dispatch) {
        return api.ListNotebooks().then(data => {
            const notebooks = data ? data : []
            dispatch(setList(notebooks))
        })
    }
}

export const createNotebook = (nb) => {
  return api.CreateNotebook(nb)
}

export const stopNotebook = (id) => {
  return api.StopNotebook(id) 
}

export const fetchResourceProfiles = () => {
  return api.ListResourceProfiles() 
}

export const updateList = notebook => {
  return {
    type: UPDATE_LIST,
    notebook
  }
}

export const setList = notebooks => {
    return {
      type: SET_LIST,
      notebooks
    }
}

  export const setFilter = filter => {
    return {
      type: SET_FILTER,
      filter
    }
}
  
  