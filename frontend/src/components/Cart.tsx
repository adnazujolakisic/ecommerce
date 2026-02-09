import { Link } from 'react-router-dom'
import { CartItem } from '../types'

interface CartProps {
  cart: CartItem[]
  updateQuantity: (productId: string, quantity: number) => void
}

export default function Cart({ cart, updateQuantity }: CartProps) {
  const total = cart.reduce((sum, item) => sum + item.price * item.quantity, 0)

  if (cart.length === 0) {
    return (
      <div className="cart-empty">
        <h2>Your cart is empty</h2>
        <p>Add some MetalBear swag to get started!</p>
        <Link to="/" className="continue-shopping-btn">Continue Shopping</Link>
      </div>
    )
  }

  return (
    <div className="cart-page">
      <h1>Shopping Cart</h1>

      <div className="cart-items">
        {cart.map(item => (
          <div key={item.productId} className="cart-item">
            <div className="cart-item-image">
              <img src={item.imageUrl} alt={item.productName} />
            </div>
            <div className="cart-item-details">
              <h3>{item.productName}</h3>
              <p className="cart-item-price">${item.price.toFixed(2)}</p>
            </div>
            <div className="cart-item-quantity">
              <button onClick={() => updateQuantity(item.productId, item.quantity - 1)}>-</button>
              <span>{item.quantity}</span>
              <button onClick={() => updateQuantity(item.productId, item.quantity + 1)}>+</button>
            </div>
            <div className="cart-item-total">
              ${(item.price * item.quantity).toFixed(2)}
            </div>
            <button
              className="remove-item-btn"
              onClick={() => updateQuantity(item.productId, 0)}
            >
              Remove
            </button>
          </div>
        ))}
      </div>

      <div className="cart-summary">
        <div className="cart-total">
          <span>Total:</span>
          <span>${total.toFixed(2)}</span>
        </div>
        <Link to="/checkout" className="checkout-btn">
          Proceed to Checkout
        </Link>
        <Link to="/" className="continue-shopping-link">
          Continue Shopping
        </Link>
      </div>
    </div>
  )
}
