import { useState, useEffect } from 'react';
import type { Task, ClientEvent } from '../../types';

interface TaskItemProps {
  task: Task;
  isActive: boolean;
  onSetActive: () => void;
  onDelete: () => void;
  sendEvent: (event: ClientEvent) => void;
}

export default function TaskItem({ task, isActive, onSetActive, onDelete, sendEvent }: TaskItemProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [headline, setHeadline] = useState(task.headline);

  useEffect(() => {
    if (!isEditing) {
      setHeadline(task.headline);
    }
  }, [task.headline, isEditing]);

  const handleUpdate = () => {
    if (headline.trim() && headline !== task.headline) {
      sendEvent({
        type: 'update_task',
        payload: {
          taskId: task.id,
          headline: headline.trim()
        }
      });
    }
    setIsEditing(false);
  };

  const estimationBadge = task.estimation ? (
    <span className="px-2 py-1 bg-green-100 text-green-800 rounded text-xs font-medium">
      {task.estimation}
    </span>
  ) : null;

  return (
    <div
      className={`
        border rounded-lg p-3
        ${isActive ? 'border-blue-500 bg-blue-50' : 'border-gray-200'}
        transition-colors cursor-pointer hover:border-blue-300
      `}
      onClick={() => !isActive && onSetActive()}
    >
      <div className="flex items-center justify-between gap-2">
        <div className="flex-1 min-w-0">
          {isEditing ? (
            <input
              type="text"
              value={headline}
              onChange={(e) => setHeadline(e.target.value)}
              onBlur={handleUpdate}
              onKeyDown={(e) => e.key === 'Enter' && handleUpdate()}
              className="w-full px-2 py-1 border rounded"
              autoFocus
              onClick={(e) => e.stopPropagation()}
            />
          ) : (
            <div className="flex items-center gap-2">
              <span className="font-medium truncate">{task.headline}</span>
              {estimationBadge}
            </div>
          )}
        </div>

        <div className="flex items-center gap-1" onClick={(e) => e.stopPropagation()}>
          {isActive && (
            <span className="text-xs bg-blue-500 text-white px-2 py-1 rounded">
              Active
            </span>
          )}
          <button
            onClick={(e) => {
              e.stopPropagation();
              setIsEditing(true);
            }}
            className="p-1 hover:bg-gray-100 rounded"
            data-testid="edit-task-button"
          >
            Edit
          </button>
          <button
            onClick={(e) => {
              e.stopPropagation();
              onDelete();
            }}
            className="p-1 hover:bg-red-100 text-red-600 rounded"
          >
            Delete
          </button>
        </div>
      </div>

      {/* Show full details only for active task */}
      {isActive && (task.description || task.trackerLink) && (
        <div className="mt-2 pt-2 border-t text-sm text-gray-600">
          {task.description && <p>{task.description}</p>}
          {task.trackerLink && (
            <a
              href={task.trackerLink}
              target="_blank"
              rel="noopener noreferrer"
              className="text-blue-600 hover:underline"
            >
              View in tracker →
            </a>
          )}
        </div>
      )}
    </div>
  );
}
