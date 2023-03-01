#
# Copyright 2023 IBM Corp.
# SPDX-License-Identifier: Apache-2.0
#
import functools
import itertools
import json

import pandas as pd
import pyarrow as pa


class Action:
    """Action is a base class for callable actions that transform record batches.

    Subclasses should:
    1. Implement `__call__` for the transformation logic
    2. Implement `schema` to calculate the transformed schema. A default implementation 
       exists so a subclass may override `field_type` and/or `field_nullable` instead.

    Args:
        description (string): textual description of this action
        columns (list): the column names this action acts on
        options (dict): additional action specific options
    """

    registry = {}

    def __init__(self, description, columns, options):
        self.metadata = {
            "name": self.__class__.__name__,
            "description": description
        }
        self.columns = columns
        self.options = options

    @property
    def name(self):
        return self.metadata["name"]

    @property
    def description(self):
        return self.metadata["description"]

    def __call__(self, records: pa.RecordBatch) -> pa.RecordBatch:
        """Transform records in a record batch."""
        return records

    def field_type(self):
        """Indicates the field type after this action runs.

        Returns:
            pyarrow.DataType: the field type or None if the action preserves the original type
        """
        return None

    def field_nullable(self):
        """Indicates the nullability of the field after this action runs.

        Returns:
            bool: boolean indicating nullability or None if the action preserves the original
        """
        return None

    def schema(self, original):
        """Computes a schema from the original schema after this action runs.

        Args:
            original (pyarrow.Schema): The original schema to be transformed

        Returns:
            pyarrow.Schema: a new schema that matches transformed data
        """
        schema: pa.Schema = original
        if not self.columns:
            self.columns = schema.names
        columns = [column for column in self.columns if column in schema.names]
        for column in columns:
            field_index = schema.get_field_index(column)
            field = schema.field(field_index)
            new_field = pa.field(field.name,
                                 self.field_type() or field.type,
                                 self.field_nullable() if self.field_nullable() is not None else field.nullable,
                                 field.metadata).with_metadata(self._build_metadata(field))
            schema = schema.set(field_index, new_field)
        return schema

    def _build_metadata(self, field: pa.Field) -> dict:
        metadata = dict(field.metadata or {})
        transformations = json.loads(metadata.get(b"transformations") or "[]")
        transformations.append(self.metadata)
        metadata[b"transformations"] = json.dumps(transformations)
        return metadata

    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        cls.registry[cls.__name__] = cls

    def __repr__(self):
        return repr(self.__dict__)


class PandasAction(Action):
    """PandasAction is a class for callable actions that transforms
       record batches by going through pandas dataframes.

    Subclasses should implement `__dftransform__` for the transformation logic
    on pandas dataframes.
    NOTE: __call__ assumes that the transformation does not change the schema
    """

    def __call__(self, records: pa.RecordBatch) -> pa.RecordBatch:
        df = records.to_pandas()
        df = self.__dftransform__(df)
        return pa.RecordBatch.from_pandas(df.dropna(how='all'), self.schema(records.schema))

    def __dftransform__(self, df: pd.DataFrame) -> pd.DataFrame:
        return df


def action_key(action):
    return str(type(action)) + str(action.options)


def consolidate_actions(actions):
    """Consolidate actions which are the same type and have the same options.

    For example, if you have two redact actions with the same redactValue
    on two different columns, consolidate them to a single action on both
    columns.

    Args:
        actions (list): list of actions to consolidate

    Returns:
        list: a new list of consolidated actions
    """

    def reduce_actions(a, b):
        a.columns = (a.columns or []) + (b.columns or [])
        return a

    actions = sorted(actions, key=action_key)
    result = []
    for k, g in itertools.groupby(actions, action_key):
        result.append(functools.reduce(reduce_actions, list(g)))
    return result


def transform(actions, record_batches):
    """A generator that applies actions on all record batches.

    Args:
        actions (list): actions to apply
        record_batches (list): list of recrod batches to act on

    Yields:
        pyarrow.RecordBatch: the next transformed record batch
    """
    for record_batch in record_batches:
        item = record_batch
        for action in actions:
            item = action(item)
        yield item


def transform_schema(actions, schema):
    """Transform schema according to actions.

    Args:
        actions (list): actions to run
        schema (pyarrow.Schema): original schema

    Returns:
        pyarrow.Schema: Transformed schema
    """
    for action in actions:
        schema = action.schema(schema)
    return schema


def transform_batches(actions, record_batches):
    """Transform record batches according to actions.

    Args:
        actions (list): actions to apply
        record_batches (list): list of recrod batches to act on

    Returns:
        [pyarrow.RecordBatch]: list of the transformed record batches
    """
    transformed_batches = []
    for record_batch in record_batches:
        item = record_batch
        for action in actions:
            item = action(item)
        transformed_batches.append(item)
    return transformed_batches
