import type { User } from '../../types';

interface UserCardProps {
  user: User;
  isCurrentUser?: boolean;
}

export default function UserCard({ user, isCurrentUser = false }: UserCardProps) {
  return (
    <div
      className={`
        flex items-center justify-between p-3 rounded-lg border transition-colors
        ${
          isCurrentUser
            ? 'bg-blue-50 border-blue-200'
            : 'bg-white border-gray-200'
        }
      `}
    >
      <div className="flex items-center gap-3">
        {/* Avatar */}
        <div
          className={`
            w-10 h-10 rounded-full flex items-center justify-center text-white font-semibold
            ${isCurrentUser ? 'bg-blue-600' : 'bg-gray-600'}
          `}
        >
          {user.name.charAt(0).toUpperCase()}
        </div>

        {/* User Info */}
        <div>
          <div className="flex items-center gap-2">
            <span className="font-medium text-gray-900">
              {user.name}
            </span>
            {isCurrentUser && (
              <span className="text-xs bg-blue-100 text-blue-700 px-2 py-0.5 rounded">
                You
              </span>
            )}
          </div>
        </div>
      </div>

      {/* Vote Status */}
      <div className="flex items-center">
        {user.isVoted ? (
          <div className="flex items-center gap-1 text-green-600">
            <svg
              className="w-5 h-5"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fillRule="evenodd"
                d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                clipRule="evenodd"
              />
            </svg>
            <span className="text-sm font-medium">Voted</span>
          </div>
        ) : (
          <div className="flex items-center gap-1 text-gray-400">
            <svg
              className="w-5 h-5"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fillRule="evenodd"
                d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z"
                clipRule="evenodd"
              />
            </svg>
            <span className="text-sm">Pending</span>
          </div>
        )}
      </div>
    </div>
  );
}
