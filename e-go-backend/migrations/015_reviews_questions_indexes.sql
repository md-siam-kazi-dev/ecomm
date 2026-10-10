-- 015_reviews_questions_indexes.sql
--
-- Additive indexes for product reviews and product questions/answers.
-- The `reviews`, `product_questions` and `product_question_answers` tables
-- already exist (migration 003); this migration only adds the lookups used by
-- the public product page and the vendor dashboard.
--
-- Applied per-statement (no transaction wrapper), matching migrations 007-014.

-- Reviews: public listing + rating aggregation per product.
CREATE INDEX IF NOT EXISTS idx_reviews_product_status
    ON reviews (product_id, status, created_at DESC);

-- Reviews: a customer's own reviews / author-scoped edits.
CREATE INDEX IF NOT EXISTS idx_reviews_user
    ON reviews (user_id);

-- Questions: public listing per product (published only).
CREATE INDEX IF NOT EXISTS idx_product_questions_product
    ON product_questions (product_id, is_published, created_at DESC);

-- Questions: a customer's own questions.
CREATE INDEX IF NOT EXISTS idx_product_questions_user
    ON product_questions (user_id);

-- Answers: grouped lookup by question.
CREATE INDEX IF NOT EXISTS idx_question_answers_question
    ON product_question_answers (question_id);

-- Vendor scoping joins (reviews/questions resolve ownership through products /
-- order_items).
CREATE INDEX IF NOT EXISTS idx_order_items_vendor
    ON order_items (vendor_id);

CREATE INDEX IF NOT EXISTS idx_order_items_product
    ON order_items (product_id);
