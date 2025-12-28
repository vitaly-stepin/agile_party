import { useState } from 'react';
import { Card, Button, Input } from '../common';
import { useTasks } from '../../context/TaskContext';
import { useWebSocket } from '../../hooks/useWebSocket';
import { useParams } from 'react-router-dom';
import TaskItem from './TaskItem';

export default function TaskList() {
  const { roomId } = useParams<{ roomId: string }>();
  const taskContext = useTasks();
  const tasks = taskContext?.tasks || [];
  const activeTask = taskContext?.activeTask || null;
  const { sendEvent } = useWebSocket(roomId || '');
  const [isCreating, setIsCreating] = useState(false);
  const [newTaskHeadline, setNewTaskHeadline] = useState('');
  const [showOnlyUnestimated, setShowOnlyUnestimated] = useState(false);

  const handleCreateTask = () => {
    if (newTaskHeadline.trim()) {
      sendEvent({
        type: 'create_task',
        payload: { headline: newTaskHeadline.trim() }
      });
      setNewTaskHeadline('');
      setIsCreating(false);
    }
  };

  const handleSetActive = (taskId: string) => {
    sendEvent({
      type: 'set_active_task',
      payload: { taskId }
    });
  };

  const handleDelete = (taskId: string) => {
    if (confirm('Are you sure you want to delete this task?')) {
      sendEvent({
        type: 'delete_task',
        payload: { taskId }
      });
    }
  };

  const estimatedCount = tasks.filter(t => t.estimation && t.estimation !== '?').length;
  const totalCount = tasks.length;

  const displayedTasks = showOnlyUnestimated
    ? tasks.filter(t => !t.estimation || t.estimation === '?')
    : tasks;

  return (
    <Card variant="outlined" padding="md">
      <div className="mb-4">
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-lg font-semibold">Tasks</h3>
          <span className="text-sm text-gray-500">
            {estimatedCount} / {totalCount} estimated
          </span>
        </div>

        <div className="flex gap-2 mb-2">
          <button
            onClick={() => setShowOnlyUnestimated(false)}
            className={`flex-1 px-3 py-1.5 text-sm rounded transition-colors ${
              !showOnlyUnestimated
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            All ({totalCount})
          </button>
          <button
            onClick={() => setShowOnlyUnestimated(true)}
            className={`flex-1 px-3 py-1.5 text-sm rounded transition-colors ${
              showOnlyUnestimated
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            Unestimated ({totalCount - estimatedCount})
          </button>
        </div>

        {!isCreating ? (
          <Button
            variant="outline"
            size="sm"
            onClick={() => setIsCreating(true)}
            fullWidth
          >
            + Add Task
          </Button>
        ) : (
          <div className="flex gap-2">
            <Input
              type="text"
              value={newTaskHeadline}
              onChange={(e) => setNewTaskHeadline(e.target.value)}
              placeholder="Task headline..."
              autoFocus
              onKeyDown={(e) => e.key === 'Enter' && handleCreateTask()}
            />
            <Button size="sm" onClick={handleCreateTask}>Add</Button>
            <Button size="sm" variant="outline" onClick={() => setIsCreating(false)}>
              Cancel
            </Button>
          </div>
        )}
      </div>

      <div className="space-y-2" data-testid="task-list">
        {displayedTasks.map(task => (
          <TaskItem
            key={task.id}
            task={task}
            isActive={task.id === activeTask?.id}
            onSetActive={() => handleSetActive(task.id)}
            onDelete={() => handleDelete(task.id)}
          />
        ))}

        {displayedTasks.length === 0 && tasks.length > 0 && (
          <div className="text-center py-8 text-gray-400">
            No {showOnlyUnestimated ? 'unestimated' : ''} tasks
          </div>
        )}

        {tasks.length === 0 && (
          <div className="text-center py-8 text-gray-400">
            No tasks yet. Add your first task to get started!
          </div>
        )}
      </div>
    </Card>
  );
}
