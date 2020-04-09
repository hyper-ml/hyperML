import makeSocket from '../../socket'


export const connectSocket = (socket, callback) => {
    return makeSocket(socket, callback)
}
 
  