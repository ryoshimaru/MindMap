import { useState } from "react";
import type { FeedbackRequest } from "../api/types";

interface FeedbackFormProps {
  isSubmitting?: boolean;
  onSubmit: (payload: FeedbackRequest) => Promise<void>;
}

export function FeedbackForm({
  isSubmitting = false,
  onSubmit
}: FeedbackFormProps) {
  const [rating, setRating] = useState(4);
  const [comment, setComment] = useState("");

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await onSubmit({
      rating,
      comment
    });
    setComment("");
    setRating(4);
  }

  return (
    <form className="form-grid" onSubmit={handleSubmit}>
      <div className="field-group">
        <label className="field-label" htmlFor="feedback-rating">
          Оценка
        </label>
        <select
          id="feedback-rating"
          className="field-control"
          value={rating}
          onChange={(event) => setRating(Number(event.target.value))}
        >
          <option value={5}>5 - отлично</option>
          <option value={4}>4 - хорошо</option>
          <option value={3}>3 - приемлемо</option>
          <option value={2}>2 - нужно доработать</option>
          <option value={1}>1 - плохо</option>
        </select>
      </div>

      <div className="field-group">
        <label className="field-label" htmlFor="feedback-comment">
          Комментарий
        </label>
        <textarea
          id="feedback-comment"
          className="field-control"
          value={comment}
          onChange={(event) => setComment(event.target.value)}
          placeholder="Что получилось хорошо и что стоит улучшить в следующей генерации?"
        />
      </div>

      <button type="submit" className="button button--secondary" disabled={isSubmitting}>
        {isSubmitting ? "Отправка..." : "Отправить отзыв"}
      </button>
    </form>
  );
}
