import { forwardRef } from 'react'
import { cn } from '../lib/utils'

interface FormFieldProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string
  error?: string
  hint?: string
}

const FormField = forwardRef<HTMLInputElement, FormFieldProps>(
  ({ label, error, hint, className, ...props }, ref) => {
    return (
      <div className="space-y-2">
        <label className="block text-sm font-display uppercase tracking-wider text-gray-400">
          {label}
          {props.required && <span className="text-neon-pink ml-1">*</span>}
        </label>
        <input
          ref={ref}
          className={cn(
            'cyber-input',
            error && 'border-neon-pink focus:border-neon-pink focus:shadow-neon-pink',
            className
          )}
          {...props}
        />
        {error && <p className="text-sm text-neon-pink font-mono">{error}</p>}
        {hint && !error && <p className="text-sm text-gray-500">{hint}</p>}
      </div>
    )
  }
)

FormField.displayName = 'FormField'

interface TextAreaFieldProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  label: string
  error?: string
  hint?: string
}

export const TextAreaField = forwardRef<HTMLTextAreaElement, TextAreaFieldProps>(
  ({ label, error, hint, className, ...props }, ref) => {
    return (
      <div className="space-y-2">
        <label className="block text-sm font-display uppercase tracking-wider text-gray-400">
          {label}
          {props.required && <span className="text-neon-pink ml-1">*</span>}
        </label>
        <textarea
          ref={ref}
          className={cn(
            'cyber-input min-h-[100px] resize-y',
            error && 'border-neon-pink focus:border-neon-pink focus:shadow-neon-pink',
            className
          )}
          {...props}
        />
        {error && <p className="text-sm text-neon-pink font-mono">{error}</p>}
        {hint && !error && <p className="text-sm text-gray-500">{hint}</p>}
      </div>
    )
  }
)

TextAreaField.displayName = 'TextAreaField'

interface SelectFieldProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  label: string
  error?: string
  hint?: string
  options: Array<{ value: string | number; label: string }>
  placeholder?: string
}

export const SelectField = forwardRef<HTMLSelectElement, SelectFieldProps>(
  ({ label, error, hint, options, placeholder, className, ...props }, ref) => {
    return (
      <div className="space-y-2">
        <label className="block text-sm font-display uppercase tracking-wider text-gray-400">
          {label}
          {props.required && <span className="text-neon-pink ml-1">*</span>}
        </label>
        <select
          ref={ref}
          className={cn(
            'cyber-input',
            error && 'border-neon-pink focus:border-neon-pink focus:shadow-neon-pink',
            className
          )}
          {...props}
        >
          {placeholder && (
            <option value="" disabled>
              {placeholder}
            </option>
          )}
          {options.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
        {error && <p className="text-sm text-neon-pink font-mono">{error}</p>}
        {hint && !error && <p className="text-sm text-gray-500">{hint}</p>}
      </div>
    )
  }
)

SelectField.displayName = 'SelectField'

export default FormField
