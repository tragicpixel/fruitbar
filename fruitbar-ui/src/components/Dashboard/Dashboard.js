import React, { useEffect, useState } from 'react';
import '../App/App.css';
import useToken from '../App/useToken';

function getList(token) {
    return fetch('http://localhost:8000/orders/', {
        headers: new Headers({
            "Access-Control-Allow-Credentials": "true",
            'Authorization': 'Bearer ' + token
        })
    })
        .then((data) => {
          return data.json();
        })
        .catch((err) => {
          console.log( err );
          throw err;
        })
}

function Dashboard() {
  const [list, setList] = useState([]);
  const { token } = useToken();

  useEffect(() => {
    let mounted = true;
    getList(token)
      .then(items => {
        if(mounted) {
          setList(items.data)
        }
      })
    return () => mounted = false;
  }, [token])

  return(
    <div className="wrapper">
     <h1>All Orders</h1>
     <ul>
       {list.map(item => <li key={item.ID}>{item.numApples}🍎 {item.numOranges}🍊 {item.numBananas}🍌 {item.numCherries}🍒 Subtotal: ${item.subtotal} Tax:${item.tax} Tip:${item.tipamt} Total:${item.total} Created: {item.CreatedAt}</li>)}
     </ul>
   </div>
  )
}

export default Dashboard;