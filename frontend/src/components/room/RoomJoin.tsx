import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

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
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="room-id">Room ID</Label>
        <Input
          id="room-id"
          type="text"
          value={roomId}
          onChange={(e) => setRoomId(e.target.value)}
          placeholder="abc12345"
          required
        />
        <p className="text-sm text-slate-500">
          Get this from the person who created the room
        </p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="join-nickname">Your Nickname</Label>
        <Input
          id="join-nickname"
          type="text"
          value={nickname}
          onChange={(e) => setNickname(e.target.value)}
          placeholder="John Doe"
          required
        />
        <p className="text-sm text-slate-500">
          This is how others will see you in the room
        </p>
      </div>

      <Button
        type="submit"
        className="w-full"
        size="lg"
        disabled={!roomId.trim() || !nickname.trim()}
      >
        Join Room
      </Button>
    </form>
  );
}
