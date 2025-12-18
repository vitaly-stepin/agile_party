import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, Input } from '../common';

export default function RoomJoin() {
  const navigate = useNavigate();
  const [roomId, setRoomId] = useState('');
  const [nickname, setNickname] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!roomId.trim() || !nickname.trim()) {
      return;
    }

    // Navigate to room with nickname in URL params
    navigate(`/room/${roomId.trim()}?nickname=${encodeURIComponent(nickname.trim())}`);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <Input
        label="Room ID"
        type="text"
        value={roomId}
        onChange={(e) => setRoomId(e.target.value)}
        placeholder="abc12345"
        required
        helperText="Get this from the person who created the room"
      />

      <Input
        label="Your Nickname"
        type="text"
        value={nickname}
        onChange={(e) => setNickname(e.target.value)}
        placeholder="John Doe"
        required
        helperText="This is how others will see you in the room"
      />

      <Button
        type="submit"
        variant="primary"
        fullWidth
        disabled={!roomId.trim() || !nickname.trim()}
      >
        Join Room
      </Button>
    </form>
  );
}
