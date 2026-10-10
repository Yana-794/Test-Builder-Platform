import { forwardRef, useId, type InputHTMLAttributes } from 'react'
import clsx from 'clsx'
import styles from './Input.module.css'

type Props = {
  label?: string
  error?: string
  hint?: string
} & InputHTMLAttributes<HTMLInputElement>

export const Input = forwardRef<HTMLInputElement, Props>(function Input(
  { label, error, hint, className, id, ...rest },
  ref
) {
  const generatedId = useId()
  const inputId = id ?? generatedId

  return (
    <div className={styles.wrap}>
      {label && (
        <label htmlFor={inputId} className={styles.label}>
          {label}
        </label>
      )}
      <input
        ref={ref}
        id={inputId}
        className={clsx(styles.input, error && styles.inputError, className)}
        {...rest}
      />
      {error ? (
        <span className={styles.error}>{error}</span>
      ) : hint ? (
        <span className={styles.hint}>{hint}</span>
      ) : null}
    </div>
  )
})