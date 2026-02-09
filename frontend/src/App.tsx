import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { useState } from 'react'
import Header from './components/Header'
import ProductList from './components/ProductList'
import ProductDetail from './components/ProductDetail'
import Cart from './components/Cart'
import Checkout from './components/Checkout'
import OrderConfirmation from './components/OrderConfirmation'
import OrderTracking from './components/OrderTracking'
import { CartItem } from './types'
import './App.css'

function App() {
  const [cart, setCart] = useState<CartItem[]>([])

  const addToCart = (item: CartItem) => {
    setCart(prev => {
      const existing = prev.find(i => i.productId === item.productId)
      if (existing) {
        return prev.map(i =>
          i.productId === item.productId
            ? { ...i, quantity: i.quantity + item.quantity }
            : i
        )
      }
      return [...prev, item]
    })
  }

  const updateQuantity = (productId: string, quantity: number) => {
    if (quantity <= 0) {
      setCart(prev => prev.filter(i => i.productId !== productId))
    } else {
      setCart(prev => prev.map(i =>
        i.productId === productId ? { ...i, quantity } : i
      ))
    }
  }

  const clearCart = () => setCart([])

  return (
    <BrowserRouter>
      <div className="app">
        <Header cartCount={cart.reduce((sum, i) => sum + i.quantity, 0)} />
        <main className="main-content">
          <Routes>
            <Route path="/" element={<ProductList addToCart={addToCart} />} />
            <Route path="/products/:id" element={<ProductDetail addToCart={addToCart} />} />
            <Route path="/cart" element={<Cart cart={cart} updateQuantity={updateQuantity} />} />
            <Route path="/checkout" element={<Checkout cart={cart} clearCart={clearCart} />} />
            <Route path="/order-confirmation/:orderNumber" element={<OrderConfirmation />} />
            <Route path="/track/:token" element={<OrderTracking />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  )
}

export default App
