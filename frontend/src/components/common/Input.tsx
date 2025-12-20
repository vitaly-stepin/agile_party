import React from 'react';

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: string;
}

export const Input: React.FC<InputProps> = ({
  label,
  error,
  helperText,
  className = '',
  id,
  ...props
}) => {
  const inputId = id || label?.toLowerCase().replace(/\s+/g, '-');

  return (
    <div className="w-full">
      {label && (
        <label htmlFor={inputId} className="block text-sm font-medium text-slate-700 mb-2">
          {label}
        </label>
      )}
      <input
        id={inputId}
        className={`
          w-full px-4 py-2.5 border rounded-lg text-slate-900 placeholder:text-slate-400
          focus:outline-none focus:ring-2 focus:ring-slate-900 focus:border-transparent
          disabled:bg-slate-50 disabled:cursor-not-allowed disabled:text-slate-500
          transition-all duration-200
          ${error ? 'border-red-500' : 'border-slate-300 hover:border-slate-400'}
          ${className}
        `}
        {...props}
      />
      {error && <p className="mt-2 text-sm text-red-600">{error}</p>}
      {helperText && !error && <p className="mt-2 text-sm text-slate-500">{helperText}</p>}
    </div>
  );
};
