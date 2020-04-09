
export const getAllNotebooks = (list, filter) => list

export const getVisibleNotebooks = (list, filter) => {
    
    if (list.length > 0) {
        return list.filter(notebook => (notebook.ID> 30))
    }

    return list
}
   