import type { VoteValue } from '../../types';

interface VoteCardProps {
  value: VoteValue;
  isSelected: boolean;
  onClick: (value: VoteValue) => void;
  disabled?: boolean;
}

export default function VoteCard({
  value,
  isSelected,
  onClick,
  disabled = false,
}: VoteCardProps) {
  return (
    <button
      onClick={() => onClick(value)}
      disabled={disabled}
      className={`
        relative aspect-[3/4] rounded-lg border-2 transition-all
        flex items-center justify-center text-3xl font-bold
        ${
          disabled
            ? 'opacity-50 cursor-not-allowed'
            : 'hover:scale-105 hover:shadow-lg cursor-pointer'
        }
        ${
          isSelected
            ? 'border-blue-600 bg-blue-50 text-blue-700 shadow-md'
            : 'border-gray-300 bg-white text-gray-700 hover:border-blue-400'
        }
      `}
    >
      {value}
    </button>
  );
}
