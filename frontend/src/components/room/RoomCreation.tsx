import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, Input } from '../common';
import { useRoom } from '../../context/RoomContext';

export default function RoomCreation() {
  const navigate = useNavigate();
  const { createRoom, error, isLoading } = useRoom();
  const [roomName, setRoomName] = useState('');
  const [nickname, setNickname] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!roomName.trim() || !nickname.trim()) {
      return;
    }

    try {
      const roomId = await createRoom(roomName.trim(), nickname.trim());
      navigate(`/room/${roomId}`);
    } catch (err) {
      // Error is handled by context
      console.error('Failed to create room:', err);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <Input
        label="Room Name"
        type="text"
        value={roomName}
        onChange={(e) => setRoomName(e.target.value)}
        placeholder="Sprint Planning #42"
        disabled={isLoading}
        required
      />

      <Input
        label="Your Nickname"
        type="text"
        value={nickname}
        onChange={(e) => setNickname(e.target.value)}
        placeholder="John Doe"
        disabled={isLoading}
        required
        helperText="This is how others will see you in the room"
      />

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded text-sm">
          {error}
        </div>
      )}

      <Button
        type="submit"
        variant="primary"
        fullWidth
        disabled={isLoading || !roomName.trim() || !nickname.trim()}
      >
        {isLoading ? 'Creating...' : 'Create Room'}
      </Button>
    </form>
  );
}
