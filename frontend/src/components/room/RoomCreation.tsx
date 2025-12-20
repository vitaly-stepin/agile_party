import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
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
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="room-name">Room Name</Label>
        <Input
          id="room-name"
          type="text"
          value={roomName}
          onChange={(e) => setRoomName(e.target.value)}
          placeholder="Sprint Planning #42"
          disabled={isLoading}
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="nickname">Your Nickname</Label>
        <Input
          id="nickname"
          type="text"
          value={nickname}
          onChange={(e) => setNickname(e.target.value)}
          placeholder="John Doe"
          disabled={isLoading}
          required
        />
        <p className="text-sm text-slate-500">
          This is how others will see you in the room
        </p>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
          {error}
        </div>
      )}

      <Button
        type="submit"
        className="w-full"
        size="lg"
        disabled={isLoading || !roomName.trim() || !nickname.trim()}
      >
        {isLoading ? (
          <>
            <Loader2 className="w-4 h-4 animate-spin" />
            Creating...
          </>
        ) : (
          'Create Room'
        )}
      </Button>
    </form>
  );
}
