import { Link } from 'react-router'
import styles from './Logo.module.css'
 import logo from '../../../assets/logo.png'

type Props = {
  to?: string;
  size?: 'sm' | 'md';
  variant?: 'default' | 'light'

}

export function Logo({ to = '/', size = 'md', variant = 'default' }: Props) {
  return (
    <Link to={to} className={`${styles.logo} ${styles[size]} ${styles[variant]}`}>
       <img src={logo} alt="" className={styles.image} />
      <span className={styles.text}>Test Builder</span>
    </Link>
  )
}