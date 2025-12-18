import { Card } from '../common';
import { useRoom } from '../../context/RoomContext';
import UserCard from './UserCard';

export default function UserList() {
  const { roomState, currentUser } = useRoom();

  const users = roomState?.users || [];
  const votedCount = users.filter((u) => u.isVoted).length;
  const totalCount = users.length;

  return (
    <Card variant="outlined" padding="lg">
      <div className="mb-4">
        <h3 className="text-lg font-semibold text-gray-900">Participants</h3>
        <p className="text-sm text-gray-600 mt-1">
          {votedCount} of {totalCount} voted
        </p>
      </div>

      {/* Progress Bar */}
      {totalCount > 0 && (
        <div className="mb-4">
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className="bg-blue-600 h-2 rounded-full transition-all duration-300"
              style={{ width: `${(votedCount / totalCount) * 100}%` }}
            />
          </div>
        </div>
      )}

      {/* User List */}
      <div className="space-y-2">
        {users.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            <p className="text-sm">No participants yet</p>
            <p className="text-xs mt-1">Share the room ID to invite others</p>
          </div>
        ) : (
          users.map((user) => (
            <UserCard
              key={user.id}
              user={user}
              isCurrentUser={currentUser?.id === user.id}
            />
          ))
        )}
      </div>
    </Card>
  );
}
