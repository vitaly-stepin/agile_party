import React, { createContext, useContext, useState, useCallback, useMemo } from 'react';
import type { ReactNode } from 'react';
import type { Task } from '../types';

interface TaskContextState {
  tasks: Task[];
  activeTask: Task | null;

  setTasks: (tasks: Task[]) => void;
  setActiveTask: (task: Task | null) => void;
  addTask: (task: Task) => void;
  updateTask: (task: Task) => void;
  removeTask: (taskId: string) => void;
  reorderTasks: (taskIds: string[]) => void;
}

const TaskContext = createContext<TaskContextState | undefined>(undefined);

interface TaskProviderProps {
  children: ReactNode;
}

export const TaskProvider: React.FC<TaskProviderProps> = ({ children }) => {
  const [tasks, setTasksInternal] = useState<Task[]>([]);
  const [activeTask, setActiveTaskInternal] = useState<Task | null>(null);

  const setTasks = useCallback((newTasks: Task[]) => {
    setTasksInternal(newTasks);
    // Active task is determined by backend logic (first unestimated)
    // Frontend just tracks it for UI highlighting
  }, []);

  const setActiveTask = useCallback((task: Task | null) => {
    setActiveTaskInternal(task);
  }, []);

  const addTask = useCallback((task: Task) => {
    setTasksInternal(prev => {
      // Check if task already exists to prevent duplicates
      if (prev.some(t => t.id === task.id)) {
        return prev;
      }
      return [...prev, task].sort((a, b) => a.position - b.position);
    });
  }, []);

  const updateTask = useCallback((updatedTask: Task) => {
    setTasksInternal(prev =>
      prev.map(t => t.id === updatedTask.id ? updatedTask : t)
    );
  }, []);

  const removeTask = useCallback((taskId: string) => {
    setTasksInternal(prev => prev.filter(t => t.id !== taskId));
    setActiveTaskInternal(prev => prev?.id === taskId ? null : prev);
  }, []);

  const reorderTasks = useCallback((taskIds: string[]) => {
    setTasksInternal(prev => {
      const taskMap = new Map(prev.map(t => [t.id, t]));
      return taskIds.map((id, index) => ({
        ...taskMap.get(id)!,
        position: index + 1
      }));
    });
  }, []);

  const value: TaskContextState = useMemo(() => ({
    tasks,
    activeTask,
    setTasks,
    setActiveTask,
    addTask,
    updateTask,
    removeTask,
    reorderTasks,
  }), [tasks, activeTask, setTasks, setActiveTask, addTask, updateTask, removeTask, reorderTasks]);

  return <TaskContext.Provider value={value}>{children}</TaskContext.Provider>;
};

export const useTasks = (): TaskContextState | undefined => {
  const context = useContext(TaskContext);
  return context;
};
