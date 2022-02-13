import React, { useState, useReducer } from 'react';
import './Product.css';

const products = [
    {
        emoji: '🍎',
        name: 'apple',
        price: 1
    },
    {
        emoji: '🍊',
        name: 'orange',
        price: 0.75
    },
    {
        emoji: '🍌',
        name: 'banana',
        price: 1.25
    },
    {
        emoji: '🍒',
        name: 'cherry',
        price: 1.5
    }
]

const currencyOptions = {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
}

function getCartItemList(cart) {
    const uniqueProductNames = getUniqueProductNames(cart)
    var itemList = "";
    uniqueProductNames.map(name => {
        itemList = itemList + " " + countProduct(cart, name) + " " + name
    })
    return itemList
}

function cartReducer(state, action) {
    switch(action.type) {
      case 'add':
        return [...state, action.product];
      case 'remove':
        const productIndex = state.findIndex(item => item.name === action.product.name);
        if(productIndex < 0) {
          return state;
        }
        const update = [...state];
        update.splice(productIndex, 1)
        return update;
      default:
        return state;
    }
}

function getTotal(cart) {
    const total = cart.reduce((totalCost, item) => totalCost + item.price, 0);
    return total.toLocaleString(undefined, currencyOptions)
}

function countProduct(cart, name) {
    const countProducts = cart.filter(product => product.name === name);
    return countProducts.length;
}

function getUniqueProductNames(cart) {
    const uniqueProductNames = [];
    cart.map(product => {
        if ( uniqueProductNames.indexOf(product.name) === -1 ) {
            uniqueProductNames.push(product.name)
        }
    })
    return uniqueProductNames
}

const formReducer = (state, event) => {
    return {
      ...state,
      [event.name]: event.value
    }
}
   


export default function Product() {
    const [submitting, setSubmitting] = useState(false);

    const [cart, setCart] = useReducer(cartReducer, []);
    const [formData, setFormData] = useReducer(formReducer, {});

    function add(product) {
        setCart({ product, type: 'add' });
    }

    function remove(product) {
        setCart({ product, type: 'remove' });
    }

    function handleSubmit(event) {
        event.preventDefault();
        alert('You have submitted the form.')
        setSubmitting(true)
    
        setTimeout(() => {
            setSubmitting(false)
        }, 3000)
    }

    function handleChange(event) {
        setFormData({
          name: event.target.name,
          value: event.target.value,
        });
      }

    return(
        <div className="wrapper">
            {submitting &&
                <div>
                    You are submitting the following:
                    <ul>
                    {Object.entries(formData).map(([name, value]) => (
                        <li key={name}><strong>{name}</strong>:{value.toString()}</li>
                    ))}
                    </ul>
                </div>
            }
            <div>
            Shopping Cart: {cart.length} total items.
            </div>
            <div>
                {getCartItemList(cart)}
            </div>
            <div>Total: {getTotal(cart)}</div>

            <div>
                <form onSubmit={handleSubmit}>
                <fieldset>
                {products.map(product => (
                    <div key={product.name}>
                        <div className="product">
                            <span role="img" aria-label={product.name}>{product.emoji}</span>
                        </div>
                        <div>Price: {product.price}</div>
                        <button type="button" onClick={() => add(product)}>Add</button>
                        <button type="button" onClick={() => remove(product)}>Remove</button>
                        <label>{product.name}:<input name={product.name} onChange={handleChange} type="number" step="1"/></label>
                    </div>
                ))}
                </fieldset>
                <button type="submit">Submit</button>
                </form>
            </div>
        </div>
        )
}