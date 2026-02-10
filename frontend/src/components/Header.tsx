import { Link } from 'react-router-dom'
import { Badge } from '@metalbear/ui'

interface HeaderProps {
  cartCount: number
}

export default function Header({ cartCount }: HeaderProps) {
  return (
    <header style={{ 
      background: 'hsl(var(--background))', 
      color: 'hsl(var(--foreground))', 
      padding: '1.375rem 0',
      minHeight: '80px',
      position: 'sticky',
      top: 0,
      zIndex: 100,
      width: '100%',
      borderBottom: '3px solid hsl(var(--foreground))',
      display: 'flex',
      alignItems: 'center'
    }}>
      <div style={{ 
        maxWidth: '1280px', 
        margin: '0 auto', 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center',
        paddingLeft: '2rem',
        paddingRight: '2rem',
        width: '100%'
      }}>
        <Link to="/" style={{ 
          display: 'flex', 
          alignItems: 'center', 
          gap: '0.75rem', 
          fontSize: '1.5rem', 
          fontWeight: 'bold',
          color: 'hsl(var(--foreground))',
          textDecoration: 'none'
        }}>
          <img src="/metal-mart.png" alt="MetalMart" style={{ height: '48px' }} />
        </Link>
        <nav style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
          <Link to="/" style={{ color: 'hsl(var(--foreground))', textDecoration: 'none', padding: '0.5rem 1rem' }}>
            Products
          </Link>
          <Link to="/cart" style={{ color: 'hsl(var(--foreground))', textDecoration: 'none', padding: '0.5rem 1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            Cart
            {cartCount > 0 && (
              <Badge style={{ backgroundColor: 'hsl(var(--primary))', color: 'hsl(var(--primary-foreground))' }}>
                {cartCount}
              </Badge>
            )}
          </Link>
        </nav>
      </div>
    </header>
  )
}
