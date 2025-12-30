import { Page, Locator } from '@playwright/test';

export class RoomPage {
  readonly page: Page;
  readonly roomTitle: Locator;
  readonly roomIdDisplay: Locator;
  readonly connectionStatus: Locator;
  readonly userList: Locator;

  // Room Management
  readonly leaveRoomButton: Locator;

  // Task Management
  readonly addTaskButton: Locator;
  readonly taskHeadlineInput: Locator;
  readonly submitTaskButton: Locator;
  readonly cancelTaskButton: Locator;
  readonly taskList: Locator;

  // Voting
  readonly votePanel: Locator;
  readonly voteStatusText: Locator;
  readonly revealVotesButton: Locator;
  readonly averageDisplay: Locator;

  constructor(page: Page) {
    this.page = page;
    this.roomTitle = page.locator('h1, h2').first();
    this.roomIdDisplay = page.locator('text=/Room ID:/i').locator('..');
    this.connectionStatus = page.getByText(/connected/i);
    this.userList = page.getByRole('list').filter({ hasText: /participants|users/i });

    // Room Management
    this.leaveRoomButton = page.getByRole('button', { name: 'Leave Room' });

    // Task Management
    this.addTaskButton = page.getByRole('button', { name: '+ Add Task' });
    this.taskHeadlineInput = page.locator('input[placeholder*="Task headline"]');
    this.submitTaskButton = page.getByRole('button', { name: 'Add' });
    this.cancelTaskButton = page.getByRole('button', { name: 'Cancel' });
    this.taskList = page.locator('[data-testid="task-list"]');

    // Voting
    this.votePanel = page.locator('[data-testid="vote-panel"]');
    this.voteStatusText = page.getByText(/You voted/);
    this.revealVotesButton = page.getByRole('button', { name: 'Reveal Votes' });
    this.averageDisplay = page.locator('.text-6xl');
  }

  async getRoomId(): Promise<string> {
    const url = this.page.url();
    const match = url.match(/\/room\/([a-f0-9-]+)/);
    if (!match) {
      throw new Error(`Could not extract room ID from URL: ${url}`);
    }
    return match[1];
  }

  async getUserInList(nickname: string): Promise<Locator> {
    return this.page.getByText(nickname);
  }

  async waitForWebSocketConnection(timeout: number = 10000) {
    await this.connectionStatus.waitFor({ state: 'visible', timeout });
  }

  async waitForUserToAppear(nickname: string, timeout: number = 10000) {
    const user = await this.getUserInList(nickname);
    await user.waitFor({ state: 'visible', timeout });
  }

  async waitForUserToDisappear(nickname: string, timeout: number = 10000) {
    const user = await this.getUserInList(nickname);
    await user.waitFor({ state: 'hidden', timeout });
  }

  // Room Management Methods
  async leaveRoom(): Promise<void> {
    await this.leaveRoomButton.click();
  }

  // Task Management Methods
  async createTask(headline: string): Promise<void> {
    await this.addTaskButton.click();
    await this.taskHeadlineInput.fill(headline);
    await this.taskHeadlineInput.press('Enter');
  }

  async clickEditTask(taskHeadline: string): Promise<void> {
    const taskItem = await this.getTaskItem(taskHeadline);
    const editButton = taskItem.getByTestId('edit-task-button');
    await editButton.click();
  }

  async updateTaskHeadline(oldHeadline: string, newHeadline: string): Promise<void> {
    // First, make sure the task is active by clicking on it
    await this.setTaskActive(oldHeadline);
    await this.page.waitForTimeout(500); // Wait for task to become active

    const taskItem = await this.getTaskItem(oldHeadline);

    // Ensure the edit button is visible and clickable
    const editButton = taskItem.getByTestId('edit-task-button');
    await editButton.waitFor({ state: 'visible', timeout: 5000 });

    // Click the edit button
    await editButton.click();

    // Wait for the input to appear
    const input = taskItem.locator('input[type="text"]');
    await input.waitFor({ state: 'visible', timeout: 5000 });
    await input.fill(newHeadline);
    await input.press('Enter');
  }

  async waitForTaskToAppear(headline: string, timeout: number = 10000): Promise<void> {
    const task = this.taskList.getByText(headline, { exact: false });
    await task.waitFor({ state: 'visible', timeout });
  }

  async getTaskItem(headline: string): Promise<Locator> {
    return this.taskList.locator('div').filter({ hasText: headline }).first();
  }

  async setTaskActive(headline: string): Promise<void> {
    const taskItem = await this.getTaskItem(headline);
    await taskItem.click();
  }

  // Voting Methods
  async selectVote(value: string): Promise<void> {
    const voteCard = this.votePanel.getByRole('button', { name: value, exact: true });
    await voteCard.click();
  }

  async clickRevealVotes(): Promise<void> {
    await this.revealVotesButton.click();
  }

  async waitForVotesRevealed(timeout: number = 10000): Promise<void> {
    await this.averageDisplay.waitFor({ state: 'visible', timeout });
  }

  async getAverageEstimate(): Promise<string> {
    return await this.averageDisplay.textContent() || '';
  }

  async isRevealButtonEnabled(): Promise<boolean> {
    return await this.revealVotesButton.isEnabled();
  }
}
