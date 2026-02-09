import { Link } from 'react-router-dom'

interface HeaderProps {
  cartCount: number
}

export default function Header({ cartCount }: HeaderProps) {
  return (
    <header className="header">
      <div className="header-content">
        <Link to="/" className="logo">
          <img src="/metal-mart.png" alt="MetalMart" className="logo-img" />
        </Link>
        <nav className="nav">
          <Link to="/" className="nav-link">Products</Link>
          <Link to="/cart" className="nav-link cart-link">
            Cart
            {cartCount > 0 && <span className="cart-badge">{cartCount}</span>}
          </Link>
        </nav>
      </div>
    </header>
  )
}
