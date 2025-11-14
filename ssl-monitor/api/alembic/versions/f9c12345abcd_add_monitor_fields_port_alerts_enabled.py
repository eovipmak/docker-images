"""Add monitor fields: port, alerts_enabled, updated_at

Revision ID: f9c12345abcd
Revises: e4b85a247cd7
Create Date: 2025-11-14 07:30:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from datetime import datetime


# revision identifiers, used by Alembic.
revision: str = 'f9c12345abcd'
down_revision: Union[str, None] = 'e4b85a247cd7'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Add new columns to monitors table
    with op.batch_alter_table('monitors', schema=None) as batch_op:
        batch_op.add_column(sa.Column('port', sa.Integer(), nullable=True))
        batch_op.add_column(sa.Column('alerts_enabled', sa.Boolean(), nullable=True))
        batch_op.add_column(sa.Column('updated_at', sa.DateTime(), nullable=True))
    
    # Set default values for existing rows
    op.execute("UPDATE monitors SET port = 443 WHERE port IS NULL")
    op.execute("UPDATE monitors SET alerts_enabled = 1 WHERE alerts_enabled IS NULL")
    op.execute(f"UPDATE monitors SET updated_at = '{datetime.utcnow().isoformat()}' WHERE updated_at IS NULL")
    
    # Make columns non-nullable after setting defaults
    with op.batch_alter_table('monitors', schema=None) as batch_op:
        batch_op.alter_column('port', nullable=False)
        batch_op.alter_column('alerts_enabled', nullable=False)
        batch_op.alter_column('updated_at', nullable=False)


def downgrade() -> None:
    # Remove added columns
    with op.batch_alter_table('monitors', schema=None) as batch_op:
        batch_op.drop_column('updated_at')
        batch_op.drop_column('alerts_enabled')
        batch_op.drop_column('port')
