import type { ButtonHTMLAttributes, ReactNode } from 'react'
import { Link } from 'react-router'
import clsx from 'clsx'
import styles from './Button.module.css'

type Variant = 'primary' | 'secondary' | 'ghost'
type Size = 'sm' | 'md' | 'lg'

type BaseProps = {
  variant?: Variant
  size?: Size
  children: ReactNode
  className?: string
}

type ButtonAsButton = BaseProps &
  ButtonHTMLAttributes<HTMLButtonElement> & {
    as?: 'button'
  }

type ButtonAsLink = BaseProps & {
  as: 'link'
  to: string
}

type Props = ButtonAsButton | ButtonAsLink

export function Button(props: Props) {
  const {
    variant = 'primary',
    size = 'md',
    children,
    className,
    ...rest
  } = props

  const classes = clsx(styles.button, styles[variant], styles[size], className)

  if (props.as === 'link') {
    return (
      <Link to={props.to} className={classes}>
        {children}
      </Link>
    )
  }

  const { as: _as, ...buttonProps } = rest as ButtonAsButton
  return (
    <button className={classes} {...buttonProps}>
      {children}
    </button>
  )
}