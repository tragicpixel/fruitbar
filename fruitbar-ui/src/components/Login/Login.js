import React, { useState } from 'react';
import PropTypes from 'prop-types';

import './Login.css';

async function loginUser(credentials) {
    return fetch('http://localhost:8001/users/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(credentials)
    })
        .then((response) => {
            if(response.ok) {
                return response.json();
            }
            throw new Error(response.status)
        })
        .catch((error) => {
            window.confirm('Login failed!')
            console.log('error: ' + error);
        })
}

export default function Login({ setToken }) {
    const [name, setUserName] = useState();
    const [password, setPassword] = useState();

    const handleSubmit = async e => {
        e.preventDefault();
        const token = await loginUser({
          name,
          password
        });
        var tokenval = token.token.replace(/"/g, "");
        setToken(tokenval);
    }

    return(
    <div className="login-wrapper">
        <h1>Please Log In</h1>
        <form onSubmit={handleSubmit}>
        <label>
            <p>Username</p>
            <input type="text" onChange={e => setUserName(e.target.value)}/>
        </label>
        <label>
            <p>Password</p>
            <input type="password" onChange={e => setPassword(e.target.value)}/>
        </label>
        <div>
            <button type="submit">Submit</button>
        </div>
        </form>
    </div>
    )
}

Login.propTypes = {
    setToken: PropTypes.func.isRequired
}