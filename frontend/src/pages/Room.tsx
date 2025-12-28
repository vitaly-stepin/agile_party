import { useEffect, useState } from 'react';
import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
import { Card, Button, Input } from '../components/common';
import { useRoom } from '../context/RoomContext';
import { useWebSocket } from '../hooks/useWebSocket';
import VotePanel from '../components/estimation/VotePanel';
import UserList from '../components/users/UserList';
import ResultsDisplay from '../components/estimation/ResultsDisplay';
import TaskList from '../components/tasks/TaskList';
import { TaskProvider } from '../context/TaskContext';

export default function Room() {
  const { roomId } = useParams<{ roomId: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { room, roomState, currentUser, joinRoom, error } = useRoom();
  const [taskDescription, setTaskDescription] = useState('');
  const [isEditingTask, setIsEditingTask] = useState(false);

  const { isConnected, sendEvent } = useWebSocket(roomId || '');

  useEffect(() => {
    if (!roomId) {
      navigate('/');
      return;
    }

    const nickname = searchParams.get('nickname');
    if (nickname) {
      joinRoom(roomId, nickname);
    }
  }, [roomId, searchParams]);

  const handleSetTask = () => {
    if (taskDescription.trim()) {
      sendEvent({
        type: 'set_task',
        payload: { description: taskDescription.trim() },
      });
      setIsEditingTask(false);
    }
  };

  const handleReveal = () => {
    sendEvent({ type: 'reveal', payload: {} });
  };

  const handleClear = () => {
    sendEvent({ type: 'clear', payload: {} });
    setTaskDescription('');
  };

  const handleLeaveRoom = () => {
    navigate('/');
  };

  const handleCopyRoomId = () => {
    if (roomId) {
      navigator.clipboard.writeText(roomId);
      // Could add a toast notification here
    }
  };

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
        <Card variant="elevated" padding="lg">
          <div className="text-center">
            <h2 className="text-xl font-semibold text-red-600 mb-2">Error</h2>
            <p className="text-gray-700 mb-4">{error}</p>
            <Button onClick={() => navigate('/')}>Back to Home</Button>
          </div>
        </Card>
      </div>
    );
  }

  if (!room || !currentUser) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading room...</p>
        </div>
      </div>
    );
  }

  return (
    <TaskProvider>
      <div className="min-h-screen bg-gray-50 pb-8">
        {/* Header */}
        <div className="bg-white border-b border-gray-200 mb-6">
          <div className="max-w-7xl mx-auto px-4 py-4">
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-2xl font-bold text-gray-900">{room.name}</h1>
                <div className="flex items-center gap-4 mt-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-gray-500">Room ID:</span>
                    <code className="text-sm font-mono bg-gray-100 px-2 py-1 rounded">
                      {roomId}
                    </code>
                    <button
                      onClick={handleCopyRoomId}
                      className="text-sm text-blue-600 hover:text-blue-700"
                    >
                      Copy
                    </button>
                  </div>
                  <div className="flex items-center gap-2">
                    <div
                      className={`w-2 h-2 rounded-full ${
                        isConnected ? 'bg-green-500' : 'bg-red-500'
                      }`}
                    />
                    <span className="text-sm text-gray-500">
                      {isConnected ? 'Connected' : 'Disconnected'}
                    </span>
                  </div>
                </div>
              </div>
              <Button variant="outline" onClick={handleLeaveRoom}>
                Leave Room
              </Button>
            </div>
          </div>
        </div>

        {/* Main Content */}
        <div className="max-w-7xl mx-auto px-4">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Left Column - Users */}
            <div className="lg:col-span-1">
              <UserList />
            </div>

            {/* Middle Column - Tasks */}
            <div className="lg:col-span-1">
              <TaskList />
            </div>

            {/* Right Column - Voting */}
            <div className="lg:col-span-1 space-y-6">
              {/* Task Description */}
              <Card variant="outlined" padding="md">
                <h3 className="text-sm font-medium text-gray-700 mb-2">
                  Current Task
                </h3>
                {isEditingTask ? (
                  <div className="flex gap-2">
                    <Input
                      type="text"
                      value={taskDescription}
                      onChange={(e) => setTaskDescription(e.target.value)}
                      placeholder="Enter task description..."
                      autoFocus
                    />
                    <Button size="sm" onClick={handleSetTask}>
                      Save
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => setIsEditingTask(false)}
                    >
                      Cancel
                    </Button>
                  </div>
                ) : (
                  <div
                    className="text-gray-900 cursor-pointer hover:text-blue-600"
                    onClick={() => setIsEditingTask(true)}
                  >
                    {roomState?.taskDescription || (
                      <span className="text-gray-400">
                        Click to add task description
                      </span>
                    )}
                  </div>
                )}
              </Card>

              {/* Results or Voting */}
              {roomState?.isRevealed ? (
                <ResultsDisplay onClear={handleClear} />
              ) : (
                <VotePanel onReveal={handleReveal} sendEvent={sendEvent} />
              )}
            </div>
          </div>
        </div>
      </div>
    </TaskProvider>
  );
}
